// make sure the structs implements the gob.GobEncoder interface
package signature

func (s *Signature) GobEncode() ([]byte, error) {
	return s.s.Serialize(), nil
}

func (s *Signature) GobDecode(data []byte) error {
	return s.s.Deserialize(data)
}

func (s *SecretKey) GobEncode() ([]byte, error) {
	return s.s.Serialize(), nil
}

func (s *SecretKey) GobDecode(data []byte) error {
	return s.s.Deserialize(data)
}

func (p *PublicKey) GobEncode() ([]byte, error) {
	return p.p.Serialize(), nil
}

func (p *PublicKey) GobDecode(data []byte) error {
	return p.p.Deserialize(data)
}
