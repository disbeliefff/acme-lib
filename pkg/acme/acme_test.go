package acme

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"github.com/go-acme/lego/registration"
	"reflect"
	"testing"
)

func TestACME_ObtainCertificate(t *testing.T) {
	type fields struct {
		Email        string
		agreeTerms   bool
		keyType      string
		privateKey   *rsa.PrivateKey
		registration *registration.Resource
	}
	type args struct {
		ctx     context.Context
		domains []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *tls.Certificate
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ACME{
				Email:        tt.fields.Email,
				agreeTerms:   tt.fields.agreeTerms,
				keyType:      tt.fields.keyType,
				privateKey:   tt.fields.privateKey,
				registration: tt.fields.registration,
			}
			got, err := a.ObtainCertificate(tt.args.ctx, tt.args.domains)
			if (err != nil) != tt.wantErr {
				t.Errorf("ObtainCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ObtainCertificate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestACME_RenewCertificate(t *testing.T) {
	type fields struct {
		Email        string
		agreeTerms   bool
		keyType      string
		privateKey   *rsa.PrivateKey
		registration *registration.Resource
	}
	type args struct {
		ctx     context.Context
		domains []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *tls.Certificate
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ACME{
				Email:        tt.fields.Email,
				agreeTerms:   tt.fields.agreeTerms,
				keyType:      tt.fields.keyType,
				privateKey:   tt.fields.privateKey,
				registration: tt.fields.registration,
			}
			got, err := a.RenewCertificate(tt.args.ctx, tt.args.domains)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenewCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RenewCertificate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewACME(t *testing.T) {
	type args struct {
		email      string
		agreeTerms bool
		keyType    string
	}
	tests := []struct {
		name    string
		args    args
		want    *ACME
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewACME(tt.args.email, tt.args.agreeTerms, tt.args.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewACME() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewACME() got = %v, want %v", got, tt.want)
			}
		})
	}
}
