package natsbackend

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/openbao/openbao/sdk/v2/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackend_Config(t *testing.T) {
	t.Run("default", func(_t *testing.T) {
		t := testBackend(_t)

		resp, err := ReadConfig(t, backendConfigEntry{})
		RequireNoRespError(t, resp, err)

		assert.Equal(t, false, resp.Data["implicit_revocations_enabled"])
	})

	t.Run("write true", func(_t *testing.T) {
		t := testBackend(_t)

		resp, err := WriteConfig(t, backendConfigEntry{}, map[string]any{
			"implicit_revocations_enabled": true,
		})
		RequireNoRespError(t, resp, err)

		resp, err = ReadConfig(t, backendConfigEntry{})
		RequireNoRespError(t, resp, err)

		assert.Equal(t, true, resp.Data["implicit_revocations_enabled"])
	})

	t.Run("write false", func(_t *testing.T) {
		t := testBackend(_t)

		resp, err := WriteConfig(t, backendConfigEntry{}, map[string]any{
			"implicit_revocations_enabled": true,
		})
		RequireNoRespError(t, resp, err)

		resp, err = UpdateConfig(t, backendConfigEntry{}, map[string]any{
			"implicit_revocations_enabled": false,
		})
		RequireNoRespError(t, resp, err)

		resp, err = ReadConfig(t, backendConfigEntry{})
		RequireNoRespError(t, resp, err)

		assert.Equal(t, false, resp.Data["implicit_revocations_enabled"])
	})

	t.Run("delete resets to default", func(_t *testing.T) {
		t := testBackend(_t)

		resp, err := WriteConfig(t, backendConfigEntry{}, map[string]any{
			"implicit_revocations_enabled": true,
		})
		RequireNoRespError(t, resp, err)

		resp, err = DeleteConfig(t, backendConfigEntry{}, nil)
		RequireNoRespError(t, resp, err)

		resp, err = ReadConfig(t, backendConfigEntry{})
		RequireNoRespError(t, resp, err)

		assert.Equal(t, false, resp.Data["implicit_revocations_enabled"])
	})
}

func TestBackend_Creds_Revoke(t *testing.T) {
	t.Run("default disabled", func(t *testing.T) {
		synctest.Test(t, func(_t *testing.T) {
			t := testBackend(_t)

			id := UserId("op1", "acc1", "user1")
			SetupTestUser(t, id, map[string]any{
				"creds_default_ttl": "1h",
			})

			sub := ReadPublicKey(t, id)
			revokeUserCreds(t, id, sub, time.Now().Add(time.Hour))

			rev, err := ReadConfig(t, id.accountId().revocationId(sub))
			RequireNoRespError(t, rev, err)
			assert.Nil(t, rev)
		})
	})

	t.Run("enabled preserves lease revocation", func(t *testing.T) {
		synctest.Test(t, func(_t *testing.T) {
			t := testBackend(_t)

			resp, err := WriteConfig(t, backendConfigEntry{}, map[string]any{
				"implicit_revocations_enabled": true,
			})
			RequireNoRespError(t, resp, err)

			id := UserId("op1", "acc1", "user1")
			SetupTestUser(t, id, map[string]any{
				"creds_default_ttl": "1h",
			})

			sub := ReadPublicKey(t, id)
			revokeUserCreds(t, id, sub, time.Now().Add(time.Hour))

			rev, err := ReadConfig(t, id.accountId().revocationId(sub))
			RequireNoRespError(t, rev, err)
			require.NotNil(t, rev)
			assert.EqualValues(t, time.Now().Unix(), rev.Data["creation_time"])
			assert.EqualValues(t, time.Hour.Seconds(), rev.Data["ttl"])
		})
	})

	t.Run("explicit revocation unaffected", func(t *testing.T) {
		synctest.Test(t, func(_t *testing.T) {
			t := testBackend(_t)

			accId := AccountId("op1", "acc1")
			SetupTestAccount(t, accId, nil)

			sub := createUserSubject(t)
			resp, err := WriteConfig(t, accId.revocationId(sub), map[string]any{
				"ttl": "10s",
			})
			RequireNoRespError(t, resp, err)

			accJwt := ReadAccountJwt(t, accId)
			assert.Contains(t, accJwt.Revocations, sub)
		})
	})
}

func revokeUserCreds(t testContext, id userId, sub string, exp time.Time) {
	t.Helper()

	b, ok := t.Backend.(*backend)
	require.True(t, ok)

	resp, err := b.userCredsRevoke(context.Background(), &logical.Request{
		Operation: logical.RevokeOperation,
		Path:      id.credsPath(),
		Storage:   t,
		Secret: &logical.Secret{
			InternalData: map[string]any{
				"op":  id.op,
				"acc": id.acc,
				"sub": sub,
				"exp": float64(exp.Unix()),
			},
		},
	}, nil)
	RequireNoRespError(t, resp, err)
}
