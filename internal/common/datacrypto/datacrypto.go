// Package datacrypto - шифрование данных
package datacrypto

import (
	"bytes"
	"crypto/rsa"
	"io"

	"github.com/4aleksei/gokeeper/internal/common/aescoder"
	"github.com/4aleksei/gokeeper/internal/common/store"
)

type DataCryptDecrypt struct {
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

func New(privKey *rsa.PrivateKey, pubKey *rsa.PublicKey) *DataCryptDecrypt {
	return &DataCryptDecrypt{
		privKey: privKey,
		pubKey:  pubKey,
	}
}

func (d *DataCryptDecrypt) Encrypt(data *store.UserData) (*store.UserDataCrypt, *aescoder.KeyAES, error) {
	key, err := aescoder.NewAES(d.pubKey)
	if err != nil {
		return nil, nil, err
	}

	dataEnc := &store.UserDataCrypt{
		Id:       data.Id,
		Uuid:     data.Uuid,
		TypeData: data.TypeData,
		EnKey:    key.GetKey(),
	}

	var wData bytes.Buffer
	wrD, err := aescoder.NewWriter(&wData, key)
	if err != nil {
		return nil, nil, err
	}
	wrD.Write([]byte(data.UserData))

	var wMe bytes.Buffer
	wrMe, err := aescoder.NewWriter(&wMe, key)
	if err != nil {
		return nil, nil, err
	}
	wrMe.Write([]byte(data.MetaData))
	dataEnc.UserDataEn = make([]byte, wData.Len())
	dataEnc.MetaDataEn = make([]byte, wMe.Len())
	copy(dataEnc.UserDataEn, wData.Bytes())
	copy(dataEnc.MetaDataEn, wMe.Bytes())
	return dataEnc, key, nil
}

func (d *DataCryptDecrypt) Decrypt(dataEnc *store.UserDataCrypt) (*store.UserData, *aescoder.KeyAES, error) {

	key, err := aescoder.DecodeAESKey(d.privKey, dataEnc.EnKey)
	if err != nil {
		return nil, nil, err
	}
	data := &store.UserData{
		Id:       dataEnc.Id,
		Uuid:     dataEnc.Uuid,
		TypeData: dataEnc.TypeData,
	}

	r := bytes.NewReader(dataEnc.UserDataEn)

	rD, err := aescoder.NewReader(io.NopCloser(r), key)
	if err != nil {
		return nil, nil, err
	}
	np, err := rD.ReadOne(dataEnc.UserDataEn)

	if err != nil {
		return nil, nil, err
	}

	data.UserData = string(np)

	rM := bytes.NewReader(dataEnc.MetaDataEn)
	rMeta, err := aescoder.NewReader(io.NopCloser(rM), key)
	if err != nil {
		return nil, nil, err
	}

	npMeta, err := rMeta.ReadOne(dataEnc.MetaDataEn)

	if err != nil {
		return nil, nil, err
	}
	data.MetaData = string(npMeta)
	return data, key, nil
}
