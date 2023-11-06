package shared

import (
	"crypto/ed25519"
	"testing"
)

func TestReadKeys(t *testing.T) {
	type args struct {
		key_path    string
		pubkey_path string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "normal",
			args:    args{"./testdata/ed25519/key.pem", "./testdata/ed25519/pubkey.pem"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priv, pub, err := ReadKeys(tt.args.key_path, tt.args.pubkey_path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			message := []byte("Hello, world!")
			signature := ed25519.Sign(priv, message)
			result := ed25519.Verify(pub, message, signature)

			if result == false {
				t.Errorf("Verification is failed.")
			}
		})
	}
}
