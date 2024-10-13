package rfccode

import (
	"BlockChainSimulator/structs"
	"crypto/ecdsa"
	"log"
	"math/big"
	"math/rand"

	"github.com/vechain/go-ecvrf"
)

type Chunk struct {
	DataBlock []byte
	Original  bool
	Padding   int
}

type CodingSchema struct {
	Round        int64
	K            int
	N            int
	RandSeedsStr string
	VRFProof     []byte
	CodingMatrix [][]byte
}

func GenRandSeed(sk ecdsa.PrivateKey, block structs.Block) (int64, []byte) {
	content := block.Hash
	vrf := ecvrf.P256Sha256Tai
	beta, pi, err := vrf.Prove(&sk, content[:])
	if err != nil {
		log.Panic(err)
	}
	randSeed := new(big.Int).SetBytes(beta).Int64()

	return randSeed, pi
}

func VerifyRandSeed(pk ecdsa.PublicKey, block structs.Block, pi []byte) bool {
	content := block.Hash
	vrf := ecvrf.P256Sha256Tai
	_, err := vrf.Verify(&pk, content[:], pi)
	if err != nil {
		return false
	}
	return true
}

func encode(chunks []Chunk, randSeed int64) (Chunk, []byte) {
	vec := make([]byte, len(chunks))
	source := rand.NewSource(randSeed)
	rander := rand.New(source)

	product := make([]byte, len(chunks[0].DataBlock))

	for i := 0; i < len(chunks); i++ {
		randNum := rander.Int()
		vec[i] = byte(randNum % FIELDSIZE)
		newProduct := blockMultiply(vec[i], chunks[i])
		product = blockSum(product, newProduct)
	}
	chunk := Chunk{
		DataBlock: product,
		Padding:   0,
		Original:  false,
	}
	return chunk, vec
}

func blockMultiply(coefficient byte, chunk Chunk) []byte {
	var product []byte
	for _, everyByte := range chunk.DataBlock {
		product = append(product, Multiply(coefficient, everyByte))
	}
	return product
}

func blockSum(product, newProduct []byte) []byte {
	var sum []byte
	for i := range newProduct {
		sum = append(sum, Add(product[i], newProduct[i]))
	}
	return sum
}

func Encode(blocks []structs.Block, randSeed int64, shardId int) (Chunk, []byte) {
	// // blocks number didn't match K
	// if len(blocks) != int(math.Log2(float64(params.K))) {
	// 	return Chunk{}, []byte{}
	// }
	// chunks := make([]Chunk, len(blocks))
	// maxLength := 0
	// for i, block := range blocks {
	// 	dataBlock := utils.Encode(block)
	// 	chunks[i].DataBlock = dataBlock
	// 	if len(dataBlock) > maxLength {
	// 		maxLength = len(dataBlock)
	// 	}
	// }

	// for i := 0; i < len(chunks); i++ {
	// 	addLength := maxLength - len(chunks[i].DataBlock)
	// 	addBytes := make([]byte, addLength)
	// 	chunks[i].DataBlock = append(chunks[i].DataBlock, addBytes...)
	// 	chunks[i].Padding = addLength
	// 	chunks[i].Original = true
	// }

	// if shardId/params.K == 0 {
	// 	vector := make([]byte, len(blocks))
	// 	vector[shardId] = 1
	// 	return chunks[shardId], vector
	// }

	// return encode(chunks, randSeed)
	return Chunk{}, []byte{}
}
