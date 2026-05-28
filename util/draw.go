package util

import (
	"bytes"
	"github.com/golang/freetype"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"net/http"
	"unicode/utf8"
	"urlAPI/file"
)

/**
 * @brief 创建带有默认字体配置的绘图上下文。
 * @return *freetype.Context 绘图上下文对象。
 */
func getDrawer() *freetype.Context {
	drawer := freetype.NewContext()
	drawer.SetDPI(144)
	drawer.SetFont(font)
	drawer.SetSrc(image.Black)
	return drawer
}

/**
 * @brief 将文本内容绘制为图片。
 * @param oriContent 原始文本内容。
 * @return []byte PNG 图片字节。
 * @return error 绘制或编码失败时返回错误。
 */
func DrawTxt(oriContent string) ([]byte, error) {
	Content := DrawTxtArrange(oriContent)

	templateImg := image.NewRGBA(image.Rect(0, 0, (25 + 40*utf8.RuneCountInString(Content[0])), (60*len(Content) + 13)))
	drawer := getDrawer()
	drawer.SetDst(templateImg)
	drawer.SetClip(templateImg.Bounds())

	drawer.SetFont(font)
	drawer.SetFontSize(25)

	for index, content := range Content {
		drawer.SetSrc(image.NewUniform(color.RGBA{100, 100, 100, 255}))
		drawer.DrawString(content, freetype.Pt(15, 60*(index+1)+2))
		drawer.SetSrc(image.White)
		drawer.DrawString(content, freetype.Pt(13, 60*(index+1)))
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, templateImg); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

/**
 * @brief 绘制仓库信息卡片。
 * @param logo 平台 Logo。
 * @param Name 仓库名称。
 * @param Author 作者名称。
 * @param Description 仓库描述。
 * @param Star Star 数。
 * @param Fork Fork 数。
 * @return []byte PNG 图片字节。
 * @return error 绘制或编码失败时返回错误。
 */
func DrawRepo(logo image.Image, Name, Author, Description, Star, Fork string) ([]byte, error) {
	starIO, _ := file.Icons.Open("icon/star_icon.png")
	forkIO, _ := file.Icons.Open("icon/fork_icon.png")
	starIcon, err := png.Decode(starIO)
	forkIcon, err := png.Decode(forkIO)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var nameLen int
	if len(Name) == len([]rune(Name)) {
		nameLen = len(Name) * 45
	} else {
		nameLen = len(Name) * 80
	}
	Author = "by " + Author
	authorLen := len(Author) * 27
	starLen := len(Star) * 27
	forkLen := len(Fork) * 27
	width := max(nameLen, authorLen) + max(starLen, forkLen) + 500

	desriptionContent := DrawWebTxtArrange(Description, width)
	height := len(desriptionContent)*50 + 300

	templateImg := image.NewRGBA(image.Rect(0, 0, width, height))
	DrawRoundedRect(templateImg, "fill")
	draw.Draw(templateImg, image.Rect(30, 30, width, height), logo, logo.Bounds().Min, draw.Over)

	drawer := getDrawer()
	drawer.SetDst(templateImg)
	drawer.SetClip(templateImg.Bounds())

	drawer.SetFontSize(50)
	drawer.DrawString(Name, freetype.Pt(260, 100))

	drawer.SetFontSize(30)
	drawer.DrawString(Author, freetype.Pt(260, 200))

	draw.Draw(templateImg, image.Rect(width-max(starLen, forkLen)-150, 30, width, height), starIcon, starIcon.Bounds().Min, draw.Over)
	draw.Draw(templateImg, image.Rect(width-max(starLen, forkLen)-150, 140, width, height), forkIcon, forkIcon.Bounds().Min, draw.Over)
	drawer.SetFontSize(30)
	drawer.DrawString(Star, freetype.Pt(width-max(starLen, forkLen)-50, 100))
	drawer.DrawString(Fork, freetype.Pt(width-max(starLen, forkLen)-50, 200))

	drawer.SetFontSize(20)
	for index, content := range desriptionContent {
		drawer.DrawString(content, freetype.Pt(30, 300+index*50))
	}

	var buf bytes.Buffer
	if err = png.Encode(&buf, templateImg); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

/**
 * @brief 绘制视频信息卡片。
 * @param CoverURL 封面图地址。
 * @param Name 视频标题。
 * @param Author 作者名称。
 * @param Description 视频描述。
 * @param View 播放量。
 * @param Favorite 收藏数。
 * @param Like 点赞数。
 * @param Coin 投币数。
 * @return []byte PNG 图片字节。
 * @return error 绘制或编码失败时返回错误。
 */
func DrawVideo(CoverURL, Name, Author, Description, View, Favorite, Like, Coin string) ([]byte, error) {
	req, err := http.NewRequest("GET", CoverURL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	pic, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pic = resize.Resize(0, 450, pic, resize.Lanczos3)

	Author = "by " + Author
	var nameLen int
	if len(Name) == len([]rune(Name)) {
		nameLen = len(Name) * 45
	} else {
		nameLen = len([]rune(Name)) * 80
	}
	authorLen := len([]rune(Author)) * 27
	statLen := (max(len(View), len(Like))+max(len(Favorite), len(Coin)))*27 + 250

	templatePic := image.NewRGBA(pic.Bounds())
	draw.Draw(templatePic, templatePic.Bounds(), pic, pic.Bounds().Min, draw.Over)
	DrawRoundedRect(templatePic, "boarder")

	width := max(nameLen, authorLen, statLen) + templatePic.Bounds().Dx() + 100
	desriptionContent := DrawWebTxtArrange(Description, width)
	height := len(desriptionContent)*50 + templatePic.Bounds().Dy() + 100

	templateImg := image.NewRGBA(image.Rect(0, 0, width, height))
	DrawRoundedRect(templateImg, "fill")
	draw.Draw(templateImg, image.Rect(30, 30, width, height), templatePic, templatePic.Bounds().Min, draw.Over)

	drawer := getDrawer()
	drawer.SetDst(templateImg)
	drawer.SetClip(templateImg.Bounds())

	drawer.SetFontSize(50)
	drawer.DrawString(Name, freetype.Pt(templatePic.Bounds().Dx()+100, 150))
	drawer.SetFontSize(30)
	drawer.DrawString(Author, freetype.Pt(templatePic.Bounds().Dx()+100, 250))

	drawer.SetFontSize(20)
	for index, content := range desriptionContent {
		drawer.DrawString(content, freetype.Pt(30, templatePic.Bounds().Dy()+index*50+100))
	}

	likeIO, _ := file.Icons.Open("icon/like_icon.png")
	favIO, _ := file.Icons.Open("icon/fav_icon.png")
	playIO, _ := file.Icons.Open("icon/play_icon.png")
	coinIO, _ := file.Icons.Open("icon/coin_icon.png")
	likeIcon, err := png.Decode(likeIO)
	favIcon, err := png.Decode(favIO)
	playIcon, err := png.Decode(playIO)
	coinIcon, err := png.Decode(coinIO)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	draw.Draw(templateImg, image.Rect(templatePic.Bounds().Dx()+100, 300, width, height), playIcon, playIcon.Bounds().Min, draw.Over)
	drawer.DrawString(View, freetype.Pt(templatePic.Bounds().Dx()+180, 350))
	draw.Draw(templateImg, image.Rect(templatePic.Bounds().Dx()+max(len(View), len(Like))*27+200, 300, width, height), favIcon, favIcon.Bounds().Min, draw.Over)
	drawer.DrawString(Favorite, freetype.Pt(templatePic.Bounds().Dx()+max(len(View), len(Like))*27+280, 350))
	draw.Draw(templateImg, image.Rect(templatePic.Bounds().Dx()+100, 400, width, height), likeIcon, likeIcon.Bounds().Min, draw.Over)
	drawer.DrawString(Like, freetype.Pt(templatePic.Bounds().Dx()+180, 450))
	draw.Draw(templateImg, image.Rect(templatePic.Bounds().Dx()+max(len(View), len(Like))*27+200, 400, width, height), coinIcon, coinIcon.Bounds().Min, draw.Over)
	drawer.DrawString(Coin, freetype.Pt(templatePic.Bounds().Dx()+max(len(View), len(Like))*27+280, 450))

	var buf bytes.Buffer
	if err = png.Encode(&buf, templateImg); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

/**
 * @brief 绘制文章信息卡片。
 * @param logo 平台 Logo。
 * @param ID 文章标识。
 * @param Title 文章标题。
 * @param Author 作者名称。
 * @param Description 摘要内容。
 * @param Time 发布时间。
 * @return []byte PNG 图片字节。
 * @return error 绘制或编码失败时返回错误。
 */
func DrawArticle(logo image.Image, ID, Title, Author, Description, Time string) ([]byte, error) {
	titleLen := len(Title) * 25
	var secondTitle string
	if Author != "" {
		secondTitle = "By " + Author
	} else {
		secondTitle = "Time: " + Time
	}
	secondLen := len(secondTitle) * 16
	width := max(titleLen, secondLen) + 60 + logo.Bounds().Dx()
	discriptionContent := DrawWebTxtArrange(Description, width)
	height := len(discriptionContent)*50 + logo.Bounds().Dy() + 100

	templateImg := image.NewRGBA(image.Rect(0, 0, width, height))
	DrawRoundedRect(templateImg, "fill")
	draw.Draw(templateImg, image.Rect(30, 30, width, height), logo, logo.Bounds().Min, draw.Over)

	drawer := getDrawer()
	drawer.SetDst(templateImg)
	drawer.SetClip(templateImg.Bounds())

	drawer.SetFontSize(15)
	drawer.DrawString(ID, freetype.Pt(60+logo.Bounds().Dx(), 50))

	drawer.SetFontSize(32)
	drawer.DrawString(Title, freetype.Pt(60+logo.Bounds().Dx(), 130))

	drawer.SetFontSize(20)
	drawer.DrawString(secondTitle, freetype.Pt(60+logo.Bounds().Dx(), 200))

	drawer.SetFontSize(20)
	for index, content := range discriptionContent {
		drawer.DrawString(content, freetype.Pt(30, 300+index*50))
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, templateImg); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

/**
 * @brief 在指定图像上绘制圆角矩形背景。
 * @param img 目标 RGBA 图像。
 * @param option 绘制模式，`fill` 为填充，其余为裁角模式。
 */
func DrawRoundedRect(img *image.RGBA, option string) {
	radius := 45
	rect := img.Bounds()

	corners := []image.Point{
		{rect.Min.X + radius, rect.Min.Y + radius},
		{rect.Max.X - radius - 1, rect.Min.Y + radius},
		{rect.Min.X + radius, rect.Max.Y - radius - 1},
		{rect.Max.X - radius - 1, rect.Max.Y - radius - 1},
	}

	if option == "fill" {
		draw.Draw(img, image.Rect(rect.Min.X+radius, rect.Min.Y, rect.Max.X-radius, rect.Max.Y), &image.Uniform{image.White}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(rect.Min.X, rect.Min.Y+radius, rect.Max.X, rect.Max.Y-radius), &image.Uniform{image.White}, image.Point{}, draw.Src)
		for _, center := range corners {
			for y := -radius; y <= radius; y++ {
				for x := -radius; x <= radius; x++ {
					if x*x+y*y <= radius*radius {
						img.Set(center.X+x, center.Y+y, image.White)
					}
				}
			}
		}
	} else {
		for x := 0; x < radius; x++ {
			for y := 0; y < radius; y++ {
				if (x-radius)*(x-radius)+(y-radius)*(y-radius) > radius*radius {
					img.Set(x, y, image.White)
					img.Set(rect.Dx()-x, y, image.White)
					img.Set(x, rect.Dy()-y, image.White)
					img.Set(rect.Dx()-x, rect.Dy()-y, image.White)
				}
			}
		}
	}
}

/**
 * @brief 将普通文本按固定宽度分行。
 * @param Str 原始文本。
 * @return []string 分行后的文本切片。
 */
func DrawTxtArrange(Str string) []string {
	Content := []rune(Str)
	var ret []string
	for i := 0; true; i += 20 {
		if i+20 >= len(Content) {
			ret = append(ret, string(Content[i:len(Content)]))
			break
		} else {
			ret = append(ret, string(Content[i:i+20]))
		}
	}
	return ret
}

/**
 * @brief 按给定宽度估算网页摘要文本的分行结果。
 * @param Str 原始文本。
 * @param Width 目标宽度。
 * @return []string 分行后的文本切片。
 */
func DrawWebTxtArrange(Str string, Width int) []string {
	var maxlen int
	Content := []rune(Str)
	if len(Str) == len(Content) {
		maxlen = (Width - 60) / 15
	} else {
		maxlen = (Width - 60) / 32
	}
	var ret []string
	for i := 0; true; i += maxlen {
		if i+maxlen >= len(Content) {
			ret = append(ret, string(Content[i:len(Content)]))
			break
		} else {
			ret = append(ret, string(Content[i:i+maxlen]))
			if (Content[i+maxlen] >= 'a' && Content[i+maxlen] <= 'z') || (Content[i+maxlen] >= 'A' && Content[i+maxlen] <= 'Z') {
				ret[len(ret)-1] += "-"
			}
		}
	}
	return ret
}
