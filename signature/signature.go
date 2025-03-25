// 对外提供一般化的签名和验证接口
package signature

import (
	"BlockChainSimulator/utils"

	"github.com/herumi/bls-go-binary/bls"
)

type Signature = bls.Sign
type SecretKey = bls.SecretKey
type PublicKey = bls.PublicKey

func init() {
	err := bls.Init(bls.BLS12_381)
	if err != nil {
		utils.LoggerInstance.Error("Failed to initialize BLS")
	}
}

// 生成密钥对
func GenerateKeyPair() (*SecretKey, *PublicKey) {
	sk := &bls.SecretKey{}
	sk.SetByCSPRNG()
	pk := sk.GetPublicKey()
	return sk, pk
}

// 生成 BLS 签名
func Sign(privateKey *SecretKey, msgHash []byte) *Signature {
	return privateKey.SignHash(msgHash)
}

// Verify 验证 BLS 签名
func Verify(publicKey *PublicKey, msgHash []byte, signature *Signature) bool {
	return signature.VerifyHash(publicKey, msgHash)
}

// 聚合签名
func AggregateSignatures(signatures []*Signature) (*Signature, error) {
	var aggregatedSignature bls.Sign
	sigs := make([]bls.Sign, len(signatures))

	for i := range signatures {
		sigs[i] = *signatures[i]
	}

	aggregatedSignature.Aggregate(sigs)

	return &aggregatedSignature, nil
}

// 验证聚合签名
func VerifyAggregatedSignature(publicKeys []*PublicKey, msgHash []byte, aggSignature *Signature) bool {
	pks := make([]bls.PublicKey, len(publicKeys))
	for i := range publicKeys {
		pks[i] = *publicKeys[i]
	}

	// 项目中聚合签名的场景是相同的消息，所以这里直接复制相同的消息
	hs := make([][]byte, len(pks))
	for i := range pks {
		hs[i] = msgHash
	}
	return aggSignature.VerifyAggregateHashes(pks, hs)
}
