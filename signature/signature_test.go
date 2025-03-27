package signature

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair(t *testing.T) {
	t.Run("should generate valid key pair", func(t *testing.T) {
		sk, pk := GenerateKeyPair()
		assert.NotNil(t, sk)
		assert.NotNil(t, pk)
		assert.False(t, sk.s.IsZero())
		assert.False(t, pk.p.IsZero())
	})

	t.Run("should generate different key pairs each time", func(t *testing.T) {
		sk1, pk1 := GenerateKeyPair()
		sk2, pk2 := GenerateKeyPair()
		assert.NotEqual(t, sk1.s.Serialize(), sk2.s.Serialize())
		assert.NotEqual(t, pk1.p.Serialize(), pk2.p.Serialize())
	})
}

func TestSignAndVerify(t *testing.T) {
	sk, pk := GenerateKeyPair()
	msg := []byte("test message")
	hash := sha256.Sum256(msg)

	t.Run("valid signature should verify", func(t *testing.T) {
		sig := Sign(sk, hash[:])
		assert.True(t, Verify(pk, hash[:], sig))
	})

	t.Run("invalid message should fail verification", func(t *testing.T) {
		sig := Sign(sk, hash[:])
		wrongHash := sha256.Sum256([]byte("wrong message"))
		assert.False(t, Verify(pk, wrongHash[:], sig))
	})

	t.Run("invalid public key should fail verification", func(t *testing.T) {
		sig := Sign(sk, hash[:])
		_, wrongPk := GenerateKeyPair()
		assert.False(t, Verify(wrongPk, hash[:], sig))
	})

	t.Run("empty message should handle gracefully", func(t *testing.T) {
		sig := Sign(sk, []byte{})
		assert.Nil(t, sig)
	})
}

func TestAggregateSignatures(t *testing.T) {
	msg := []byte("aggregate test")
	hash := sha256.Sum256(msg)

	t.Run("should aggregate single signature", func(t *testing.T) {
		sk, pk := GenerateKeyPair()
		sig := Sign(sk, hash[:])

		aggSig, err := AggregateSignatures([]*Signature{sig})
		require.NoError(t, err)
		assert.True(t, VerifyAggregatedSignature([]*PublicKey{pk}, hash[:], aggSig))
	})

	t.Run("should aggregate multiple signatures", func(t *testing.T) {
		const n = 5
		sks := make([]*SecretKey, n)
		pks := make([]*PublicKey, n)
		sigs := make([]*Signature, n)

		for i := 0; i < n; i++ {
			sks[i], pks[i] = GenerateKeyPair()
			sigs[i] = Sign(sks[i], hash[:])
		}

		aggSig, err := AggregateSignatures(sigs)
		require.NoError(t, err)
		assert.True(t, VerifyAggregatedSignature(pks, hash[:], aggSig))
	})

	t.Run("empty signatures should return error", func(t *testing.T) {
		_, err := AggregateSignatures([]*Signature{})
		assert.Error(t, err)
	})

	t.Run("should detect invalid aggregated signature", func(t *testing.T) {
		sk1, _ := GenerateKeyPair()
		_, pk2 := GenerateKeyPair() // different key not used for signing
		sig := Sign(sk1, hash[:])

		aggSig, err := AggregateSignatures([]*Signature{sig})
		require.NoError(t, err)
		assert.False(t, VerifyAggregatedSignature([]*PublicKey{pk2}, hash[:], aggSig))
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("nil secret key should return nil on Sign", func(t *testing.T) {
		msg := []byte("test")
		hash := sha256.Sum256(msg)
		assert.Nil(t, Sign(nil, hash[:]))
	})

	t.Run("nil public key should return false on verify", func(t *testing.T) {
		sk, _ := GenerateKeyPair()
		msg := []byte("test")
		hash := sha256.Sum256(msg)
		sig := Sign(sk, hash[:])
		assert.False(t, Verify(nil, hash[:], sig))
	})

	t.Run("nil signature should return false on verify", func(t *testing.T) {
		_, pk := GenerateKeyPair()
		msg := []byte("test")
		hash := sha256.Sum256(msg)
		assert.False(t, Verify(pk, hash[:], nil))
	})
}

func BenchmarkSign(b *testing.B) {
	sk, _ := GenerateKeyPair()
	msg := []byte("benchmark message")
	hash := sha256.Sum256(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Sign(sk, hash[:])
	}
}

func BenchmarkVerify(b *testing.B) {
	sk, pk := GenerateKeyPair()
	msg := []byte("benchmark message")
	hash := sha256.Sum256(msg)
	sig := Sign(sk, hash[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Verify(pk, hash[:], sig)
	}
}
