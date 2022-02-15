package ldappool

import "github.com/go-ldap/ldap/v3"

func (p *Pool) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	pc, err := p.pull()
	if err != nil {
		return nil, err
	}

	res, err := pc.Search(searchRequest)
	defer p.release(pc)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *Pool) SearchWithPaging(searchRequest *ldap.SearchRequest, pagingSize uint32) (*ldap.SearchResult, error) {
	pc, err := p.pull()
	if err != nil {
		return nil, err
	}

	res, err := pc.SearchWithPaging(searchRequest, pagingSize)
	defer p.release(pc)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *Pool) Add(addRequest *ldap.AddRequest) error {
	pc, err := p.pull()
	if err != nil {
		return err
	}

	err = pc.Add(addRequest)
	defer p.release(pc)

	if err != nil {
		return err
	}

	return nil
}

func (p *Pool) Modify(modifyRequest *ldap.ModifyRequest) error {
	pc, err := p.pull()
	if err != nil {
		return err
	}

	err = pc.Modify(modifyRequest)
	defer p.release(pc)

	if err != nil {
		return err
	}

	return nil
}

func (p *Pool) ModifyDN(m *ldap.ModifyDNRequest) error {
	pc, err := p.pull()
	if err != nil {
		return err
	}

	err = pc.ModifyDN(m)
	defer p.release(pc)

	if err != nil {
		return err
	}

	return nil
}

func (p *Pool) ModifyWithResult(modifyRequest *ldap.ModifyRequest) (*ldap.ModifyResult, error) {
	pc, err := p.pull()
	if err != nil {
		return nil, err
	}

	res, err := pc.ModifyWithResult(modifyRequest)
	defer p.release(pc)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *Pool) Del(delRequest *ldap.DelRequest) error {
	pc, err := p.pull()
	if err != nil {
		return err
	}

	err = pc.Del(delRequest)
	defer p.release(pc)

	if err != nil {
		return err
	}

	return nil
}
