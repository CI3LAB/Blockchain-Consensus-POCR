package message

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/big"

	"github.com/yoseplee/vrf"
	"golang.org/x/crypto/ed25519"
	"gonum.org/v1/gonum/stat/distuv"
)

func Hash(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(content)
}

func Digest(obj interface{}) (string, error) {
	content, err := json.Marshal(obj)
	if err != nil {
		log.Printf("[Crypto] marshl the object error: %s", err)
		return "", err
	}
	return Hash(content), nil
}

func GetPriPubKey() ([]byte, []byte, error) {
	pub, pri, err := ed25519.GenerateKey(rand.Reader)
	return pub, pri, err
}

func ByteToString(bytevalue []byte) string {
	str := base64.StdEncoding.EncodeToString(bytevalue)
	return str
}

func StringToByte(strvalue string) []byte {
	bytevalue, _ := base64.StdEncoding.DecodeString(strvalue)
	return bytevalue
}

func GenerateProofAndHash(pk []byte, sk []byte, m []byte) ([]byte, []byte, error) {
	proof, hash, err := vrf.Prove(pk, sk, m)
	return proof, hash, err
}

func VerifyVRF(pk []byte, proof []byte, m []byte) (bool, error) {
	boolvalue, err := vrf.Verify(pk, proof, m)
	return boolvalue, err
}

func CalculateRatio(hash []byte) *big.Float {
	hashlen := len(hash)
	hashInt := new(big.Int).SetBytes(hash)
	denominator := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(hashlen)*8), nil)
	hashFloat := new(big.Float).SetInt(hashInt)
	denominatorFloat := new(big.Float).SetInt(denominator)
	ratio := new(big.Float).Quo(hashFloat, denominatorFloat)
	return ratio
}

func CalculateJ(nodeCRcredit float64, hashratio *big.Float) int {
	p := 0.3 // set according to committee size
	hashratiofloat, _ := hashratio.Float64()
	jValue := 0
	for float64(jValue) <= nodeCRcredit {
		jBinomialCDF := GetCDFBinomial(float64(jValue), nodeCRcredit, p)
		if hashratiofloat > jBinomialCDF {
			jValue++
		} else {
			break
		}
	}
	return jValue
}

func VerififyJ(totalCRCredit float64, nodeCRcredit float64, hashratio *big.Float, jValue int) bool {
	p := 0.5
	hashratiofloat, _ := hashratio.Float64()
	jBinomialCDF := GetCDFBinomial(float64(jValue), nodeCRcredit, p)
	mBinomialCDF := GetCDFBinomial(float64(jValue-1), nodeCRcredit, p)
	if (hashratiofloat < mBinomialCDF) || (hashratiofloat >= jBinomialCDF) {
		return false
	}
	return true
}

func GetCDFBinomial(x float64, n float64, p float64) float64 {
	dist := distuv.Binomial{
		N:   n,
		P:   p,
		Src: nil,
	}
	return dist.CDF(x)
}
