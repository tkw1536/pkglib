//spellchecker:words sshx
package sshx_test

//spellchecker:words crypto ecdsa elliptic rand reflect testing github pkglib sshx golang
import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"reflect"
	"testing"

	"go.tkw01536.de/pkglib/sshx"
	"golang.org/x/crypto/ssh"
)

type TestKey struct {
	Key     ssh.PublicKey
	Encoded []byte
}

var testKeys []TestKey
var testsKeysSerialized []byte

func initKey[K any](key K, err error) (tk TestKey) {
	if err != nil {
		panic(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		panic(err)
	}

	tk.Key = signer.PublicKey()
	tk.Encoded = ssh.MarshalAuthorizedKey(tk.Key)
	return tk
}

func init() {
	testKeys = []TestKey{
		initKey(rsa.GenerateKey(rand.Reader, 2048)),
		initKey(rsa.GenerateKey(rand.Reader, 2048)),
		initKey(ecdsa.GenerateKey(elliptic.P256(), rand.Reader)),
		initKey(ecdsa.GenerateKey(elliptic.P256(), rand.Reader)),
	}

	// turn all the serialized keys into one serialization
	for _, key := range testKeys {
		testsKeysSerialized = append(testsKeysSerialized, key.Encoded...)
	}
}

func TestParseKeys(t *testing.T) {
	t.Parallel()

	type args struct {
		in    []byte
		limit int
	}
	tests := []struct {
		name         string
		args         args
		wantKeys     []ssh.PublicKey
		wantComments []string
		wantOptions  [][]string
		wantRest     []byte
		wantErr      bool
	}{
		{
			name: "parse all keys",
			args: args{in: testsKeysSerialized, limit: -1},
			wantKeys: []ssh.PublicKey{
				testKeys[0].Key, testKeys[1].Key, testKeys[2].Key, testKeys[3].Key,
			},
			wantOptions:  [][]string{nil, nil, nil, nil},
			wantComments: []string{"", "", "", ""},
			wantRest:     nil,
			wantErr:      false,
		},
		{
			name: "parse only 1 key",
			args: args{in: testsKeysSerialized, limit: 1},
			wantKeys: []ssh.PublicKey{
				testKeys[0].Key,
			},
			wantOptions:  [][]string{nil},
			wantComments: []string{""},
			wantRest:     testsKeysSerialized[len(testKeys[0].Encoded):],
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotKeys, gotComments, gotOptions, gotRest, err := sshx.ParseKeys(tt.args.in, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("ParseKeys() gotKeys = %v, want %v", gotKeys, tt.wantKeys)
			}
			if !reflect.DeepEqual(gotComments, tt.wantComments) {
				t.Errorf("ParseKeys() gotComments = %v, want %v", gotComments, tt.wantComments)
			}
			if !reflect.DeepEqual(gotOptions, tt.wantOptions) {
				t.Errorf("ParseKeys() gotOptions = %v, want %v", gotOptions, tt.wantOptions)
			}
			if !reflect.DeepEqual(gotRest, tt.wantRest) {
				t.Errorf("ParseKeys() gotRest = %v, want %v", gotRest, tt.wantRest)
			}
		})
	}
}

func TestParseAllKeys(t *testing.T) {
	t.Parallel()

	type args struct {
		in []byte
	}
	tests := []struct {
		name     string
		args     args
		wantKeys []ssh.PublicKey
	}{
		{
			name: "parse all keys",
			args: args{
				in: testsKeysSerialized,
			},
			wantKeys: []ssh.PublicKey{
				testKeys[0].Key, testKeys[1].Key, testKeys[2].Key, testKeys[3].Key,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if gotKeys := sshx.ParseAllKeys(tt.args.in); !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("ParseAllKeys() = %v, want %v", gotKeys, tt.wantKeys)
			}
		})
	}
}
