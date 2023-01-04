package ldappool

import (
	"context"
	"time"

	"github.com/m-vinc/ldap/v3"
)

func (p *Pool) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
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

func (p *Pool) PasswordModify(passwordModifyRequest *ldap.PasswordModifyRequest) (*ldap.PasswordModifyResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
	if err != nil {
		return nil, err
	}

	res, err := pc.PasswordModify(passwordModifyRequest)
	defer p.release(pc)

	if err != nil {
		return nil, err
	}

	return res, err
}

func (p *Pool) Add(addRequest *ldap.AddRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pc, err := p.pull(ctx)
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
