package main

import (
	c_rand "crypto/rand"
	"embed"
	"io/fs"
	"math/big"
	"math/rand"
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
)

//go:embed assets
var embededFiles embed.FS

var (
	PRIME *big.Int
)

func init() {
	PRIME, _ = c_rand.Prime(c_rand.Reader, 256)
}

type PointJson struct {
	X int    `json:"x"`
	Y string `json:"y"`
}

func (json *PointJson) ToPoint() Point {
	x := json.X
	y, _ := new(big.Int).SetString(json.Y, 10)
	return Point{
		X: x,
		Y: y,
	}
}

type Point struct {
	X int
	Y *big.Int
}

func (p *Point) ToJson() PointJson {
	return PointJson{
		X: p.X,
		Y: p.Y.String(),
	}
}

func main() {
	app := iris.New()
	fsys, _ := fs.Sub(embededFiles, "assets")
	app.Get("/", iris.FromStd(http.FileServer(http.FS(fsys))))
	app.Post("/generate", generate)
	app.Post("/decrypt", decrypt)
	app.Listen(":8080")
}

func generate(ctx iris.Context) {
	type generateRequest struct {
		T int `json:"t"`
		N int `json:"n"`
	}
	type generateResponse struct {
		Secret string      `json:"secret"`
		Points []PointJson `json:"points"`
	}

	form := &generateRequest{}
	if err := ctx.ReadForm(form); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		ctx.StopExecution()
		return
	}
	if form.T <= 0 || form.N <= 0 || form.T > form.N {
		ctx.StatusCode(iris.StatusUnprocessableEntity)
		ctx.WriteString("t must be between 1 and n")
		ctx.StopExecution()
		return
	}
	secret, points := generateRandomShares(form.T, form.N, PRIME)
	jsons := TransSlice(points, func(p Point) PointJson { return p.ToJson() })
	ctx.JSON(generateResponse{
		Secret: secret.String(),
		Points: jsons,
	})
	ctx.Next()
}

func generateRandomShares(t, n int, p *big.Int) (secret *big.Int, points []Point) {
	poly := make([]*big.Int, t)
	for i := 0; i < t; i++ {
		poly[i] = new(big.Int).Rand(rand.New(rand.NewSource(time.Now().Unix())), p)
	}
	secret = poly[0]

	termAt := func(n int) *big.Int {
		ans := big.NewInt(0)
		for i, coe := range poly {
			x := pow(big.NewInt(int64(n)), big.NewInt(int64(i)), p)
			x.Mul(x, coe)
			ans.Add(ans, x).Mod(ans, p)
		}
		return ans
	}

	for i := 0; i < n; i++ {
		points = append(points, Point{X: i + 1, Y: termAt(i + 1)})
	}
	return
}

func decrypt(ctx iris.Context) {
	type decryptRequest struct {
		Points []PointJson `json:"points"`
	}
	type decryptResponse struct {
		DecryptedSecret string `json:"decrypted_secret"`
	}
	json := &decryptRequest{}
	if err := ctx.ReadJSON(json); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		ctx.StopExecution()
		return
	}
	if len(json.Points) < 1 {
		ctx.StatusCode(iris.StatusUnprocessableEntity)
		ctx.WriteString("points must be at least 1")
		ctx.StopExecution()
		return
	}
	points := TransSlice(json.Points, func(p PointJson) Point { return p.ToPoint() })
	secret := lagrangeInterpolation(points, PRIME)
	ctx.JSON(decryptResponse{
		DecryptedSecret: secret.String(),
	})
	ctx.Next()
}

func lagrangeInterpolation(points []Point, p *big.Int) *big.Int {
	ans := big.NewInt(0)
	for i, point := range points {
		up := big.NewInt(1)
		down := big.NewInt(1)
		for j, otherPoint := range points {
			if i == j {
				continue
			}
			up.Mul(up, big.NewInt(int64(-otherPoint.X)))
			down.Mul(down, big.NewInt(int64(point.X-otherPoint.X)))
		}
		// 费马小定理求逆元
		term := new(big.Int)
		term.Mul(up, pow(down, new(big.Int).Sub(p, big.NewInt(2)), p)).Mod(term, p)
		term.Mul(term, point.Y).Mod(term, p)
		ans.Add(ans, term).Mod(ans, p)
	}
	return ans
}

// utils

func pow(a, b, p *big.Int) *big.Int {
	ans := big.NewInt(1)
	base := new(big.Int).Set(a)
	for i := 0; i < b.BitLen(); i++ {
		if b.Bit(i) == 1 {
			ans.Mul(ans, base).Mod(ans, p)
		}
		base.Mul(base, base).Mod(base, p)
	}
	return ans
}

func TransSlice[T, U any](s []T, trans func(T) U) (us []U) {
	for _, t := range s {
		us = append(us, trans(t))
	}
	return
}
