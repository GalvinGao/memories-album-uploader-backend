package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/upyun/go-sdk/upyun"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Validator struct {
	validator *validator.Validate
}

func (cv *Validator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var DB *gorm.DB
var Up *upyun.UpYun

func main() {
	fmt.Println("Starting memories-album-backend Service...")

	e := echo.New()

	fmt.Println("Initiating DB Connection...")

	var err error
	DB, err = gorm.Open("mysql", "username:password@tcp(localhost:3306)/memoriesalbum?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}

	DB.LogMode(true)

	e.Use(middleware.CORS())
	//e.Use(middleware.HTTPSRedirect())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())
	//e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
	//	Generator: func() string {
	//		return uuid.Must(uuid.NewV4()).String()
	//	},
	//}))

	e.Validator = &Validator{
		validator: validator.New(),
	}

	fmt.Println("Initiating Upyun Service...")

	Up := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   UpyunBucket,
		Operator: UpyunOperator,
		Password: UpyunPassword,
	})

	api := e.Group("/api")

	api.GET("/merged", func(c echo.Context) error {
		faces, err := getFaces()
		if err != nil {
			log.Printf("find faces error: %v", err)
			return DefaultServerErrorResponse
		}

		people, err := getPeople()
		if err != nil {
			log.Printf("find people error: %v", err)
			return DefaultServerErrorResponse
		}

		images, err := getImages()
		if err != nil {
			log.Printf("find images error: %v", err)
			return DefaultServerErrorResponse
		}

		return c.JSON(http.StatusOK, &MergedResponse{
			Faces:  faces,
			People: people,
			Images: images,
		})
	})

	api.POST("/upload/initiate", func(c echo.Context) error {
		var request UploadRequest
		if err := c.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		if DB.First(&Person{ID: request.PersonID}).Error != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad Request: invalid `peopleId` provided")
		}

		ext := strings.ToLower(strings.TrimPrefix(path.Ext(request.Filename), "."))
		if !contains(UpyunUploadPermittedFileTypes, ext) {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad Request: not allowed file type")
		}

		filename := fmt.Sprintf("%s.%s", uuid.Must(uuid.NewV4()).String(), ext)

		saveKey := fmt.Sprintf("%s%s", UpyunPrefix, filename)

		apps := struct {
			Name string `json:"name"`
			Params string `json:"x-gmkerl-thumb"`
			SaveAs string `json:"save_as"`
			NotifyURL string `json:"notify_url"`
		}{
			Name: "thumb",
			Params: UpyunUploadPostProcessingParams,
			SaveAs: fmt.Sprintf("%s%s", UpyunThumbPrefix, filename),
			NotifyURL: fmt.Sprintf("https://example.com"),
		}

		if err != nil {
			return DefaultServerErrorResponse
		}

		policyMap := map[string]interface{}{
			"bucket": UpyunBucket,
			"save-key": saveKey,
			"expiration": strconv.FormatInt(time.Now().Add(time.Minute * UpyunTokenExpiration).Unix(), 10),
			"allow-file-type": strings.Join(UpyunUploadPermittedFileTypes, ","),
			"content-length-range": UpyunUploadContentLengthRange,
			"image-height-range": UpyunUploadImageHeightRange,
			"image-width-range": UpyunUploadImageWidthRange,
			"x-gmkerl-thumb": UpyunUploadSyncProcessingParams,
			"notify-url": fmt.Sprintf("%s%s", Host, "/api/upload/callback/task"),
			"ext-param": request.PersonID,
			"apps": []interface{}{apps},
		}
		policyBytes, err := json.Marshal(policyMap)
		if err != nil {
			return DefaultServerErrorResponse
		}

		policy := base64.StdEncoding.EncodeToString(policyBytes)
		auth := Up.MakeUnifiedAuth(&upyun.UnifiedAuthConfig{
			Method:     "POST",
			Uri:        fmt.Sprintf("/%s", UpyunBucket),
			Policy:     policy,
		})
		return c.JSON(http.StatusOK, UploadInitiateResponse{
			Bucket: UpyunBucket,
			Authorization: auth,
			Policy: policy,
			Filename: filename,
		})
	})

	api.POST("/upload/callback", func(c echo.Context) error {
		var request UploadCallbackRequest
		if err := c.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		request.PersonID = strings.TrimSpace(request.PersonID)

		if DB.First(&Person{ID: request.PersonID}).Error != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad Request: invalid `ext-param` provided")
		}

		fmt.Println(
			fmt.Sprintf(
				"Such File has been uploaded successfully to Upyun (callback): %v",
				spew.Sdump(request)))

		t := time.Unix(request.Time, 0)

		err := DB.Transaction(func(tx *gorm.DB) error {
			image := Image{
				CdnLocation: path.Base(request.URL),
				Height:      request.ImageHeight,
				Width:       request.ImageWidth,
				Time:        t.Format(ImageTimeFormat),
				Source:      UploadImageType,
			}

			if err := tx.Create(&image).Error; err != nil {
				log.Println("failed to create new image: ", err)
				return err
			}

			if err := tx.Create(&Face{
				ID:             uuid.Must(uuid.NewV4()).String(),
				ParentPersonID: request.PersonID,
				ParentImageID:  image.ID,
			}).Error; err != nil {
				log.Println("failed to create new image: ", err)
				return err
			}

			return nil
		})

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusAccepted)
	})

	api.POST("/upload/callback/task", func(c echo.Context) error {
		return c.NoContent(http.StatusAccepted)
	})

	fmt.Println("Starting HTTP Service...")

	log.Fatal(e.Start("localhost:8000"))
}
