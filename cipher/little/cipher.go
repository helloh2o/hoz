package little

type LittleCipher struct {
	// 编码用的密码
	encodePassword *password
	// 解码用的密码
	decodePassword *password
}

// 加密原数据
func (cipher *LittleCipher) Encrypt(bs []byte) ([]byte, error) {
	for i, v := range bs {
		bs[i] = cipher.encodePassword[v]
	}
	return bs, nil
}

// 解码加密后的数据到原数据
func (cipher *LittleCipher) Decrypt(bs []byte) ([]byte, error) {
	for i, v := range bs {
		bs[i] = cipher.decodePassword[v]
	}
	return bs, nil
}

// 新建一个编码解码器
func NewCipher(encodePassword *password) *LittleCipher {
	decodePassword := &password{}
	for i, v := range encodePassword {
		encodePassword[i] = v
		decodePassword[v] = byte(i)
	}
	return &LittleCipher{
		encodePassword: encodePassword,
		decodePassword: decodePassword,
	}
}
