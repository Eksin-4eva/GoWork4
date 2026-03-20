package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/BiliGO/biz/dal/model"
	"github.com/BiliGO/biz/dal/mysql"
	"github.com/BiliGO/biz/dal/query"
	"github.com/pquerna/otp/totp"
)

type MFAQrcodeResult struct {
	Secret string
	Qrcode string
}

func GenerateMFAQrcode(ctx context.Context, userID int64, username string) (*MFAQrcodeResult, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "BiliGO",
		AccountName: username,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp key failed: %w", err)
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("generate qrcode image failed: %w", err)
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode qrcode image failed: %w", err)
	}

	qrcodeBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	q := query.Use(mysql.DB)
	u := q.User
	_, err = u.WithContext(ctx).Where(u.ID.Eq(userID)).UpdateSimple(u.MFASecret.Value(key.Secret()))
	if err != nil {
		return nil, fmt.Errorf("save mfa secret failed: %w", err)
	}

	return &MFAQrcodeResult{
		Secret: key.Secret(),
		Qrcode: qrcodeBase64,
	}, nil
}

func VerifyMFA(ctx context.Context, userID int64, code string) (bool, error) {
	q := query.Use(mysql.DB)
	u := q.User

	user, err := u.WithContext(ctx).Where(u.ID.Eq(userID)).First()
	if err != nil {
		return false, err
	}

	if user.MFASecret == "" {
		return false, fmt.Errorf("mfa not enabled")
	}

	valid := totp.Validate(code, user.MFASecret)
	return valid, nil
}

func GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	q := query.Use(mysql.DB)
	u := q.User
	return u.WithContext(ctx).Where(u.ID.Eq(userID)).First()
}

func BindMFA(ctx context.Context, userID int64, code, secret string) error {
	valid := totp.Validate(code, secret)
	if !valid {
		return fmt.Errorf("invalid mfa code")
	}

	q := query.Use(mysql.DB)
	u := q.User
	_, err := u.WithContext(ctx).Where(u.ID.Eq(userID)).UpdateSimple(u.MFASecret.Value(secret))
	if err != nil {
		return fmt.Errorf("save mfa secret failed: %w", err)
	}

	return nil
}
