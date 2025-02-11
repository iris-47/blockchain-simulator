// 对外提供一般化的签名和验证接口
package signature

import (
	"BlockChainSimulator/utils"

	"github.com/herumi/bls-go-binary/bls"
)

func init() {
	err := bls.Init(bls.BLS12_381)
	if err != nil {
		utils.LoggerInstance.Error("Failed to initialize BLS")
	}
}

// 生成密钥对
func GenerateKeyPair() ([]byte, []byte) {
	sk := bls.SecretKey{}
	sk.SetByCSPRNG()
	pk := sk.GetPublicKey()
	return sk.Serialize(), pk.Serialize()
}

// 生成 BLS 签名
func Sign(privateKey []byte, msgHash []byte) ([]byte, error) {
	var sk bls.SecretKey
	err := sk.Deserialize(privateKey)

	if err != nil {
		utils.LoggerInstance.Error("Failed to deserialize the private key")
	}

	sig := sk.SignHash(msgHash)
	return sig.Serialize(), nil
}

// Verify 验证 BLS 签名
func Verify(publicKey []byte, msgHash []byte, signature []byte) bool {
	var pk bls.PublicKey
	err := pk.Deserialize(publicKey)
	if err != nil {
		utils.LoggerInstance.Error("Failed to deserialize the public key")
	}

	var sig bls.Sign
	err = sig.Deserialize(signature)
	if err != nil {
		utils.LoggerInstance.Error("Failed to deserialize the signature")
	}
	return sig.VerifyHash(&pk, msgHash)
}

// 聚合签名
func AggregateSignatures(signatures [][]byte) ([]byte, error) {
	var aggregatedSignature bls.Sign
	sigs := make([]bls.Sign, len(signatures))

	for i := range sigs {
		var sig bls.Sign
		err := sig.Deserialize(signatures[i])
		if err != nil {
			utils.LoggerInstance.Error("Failed to deserialize the signature")
		}
		sigs[i] = sig
	}

	aggregatedSignature.Aggregate(sigs)

	return aggregatedSignature.Serialize(), nil
}

// 验证聚合签名
func VerifyAggregatedSignature(publicKeys [][]byte, msgHash []byte, signature []byte) bool {
	var aggregatedSignature bls.Sign
	err := aggregatedSignature.Deserialize(signature)
	if err != nil {
		utils.LoggerInstance.Error("Failed to deserialize the signature")
	}

	pks := make([]bls.PublicKey, len(publicKeys))
	for i := range publicKeys {
		var pk bls.PublicKey
		err := pk.Deserialize(publicKeys[i])
		if err != nil {
			utils.LoggerInstance.Error("Failed to deserialize the public key")
		}
		pks[i] = pk
	}

	// 项目中聚合签名的场景是相同的消息，所以这里直接复制相同的消息
	hs := make([][]byte, len(pks))
	for i := range pks {
		hs[i] = msgHash
	}
	return aggregatedSignature.VerifyAggregateHashes(pks, hs)
}
