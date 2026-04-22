package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"honoka-chan/config"
	"honoka-chan/encrypt"
	"honoka-chan/model"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

type PkgInfo struct {
	Id    int `xorm:"pkg_id"`
	Order int `xorm:"pkg_order"`
	Size  int `xorm:"pkg_size"`
}

func logSampleURLs(tag string, urls []string, max int) {
	if len(urls) == 0 {
		log.Printf("[DL] %s urls=0", tag)
		return
	}
	if len(urls) > max {
		urls = urls[:max]
	}
	log.Printf("[DL] %s urls(sample)=%v", tag, urls)
}

func remoteSize(url string) int {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0
	}
	cl := resp.Header.Get("Content-Length")
	if cl == "" {
		return 0
	}
	n, err := strconv.ParseInt(cl, 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return int(n)
}

func pkgSizeOrRemote(pkgSize int, url string) int {
	if pkgSize > 0 {
		return pkgSize
	}
	size := remoteSize(url)
	log.Printf("[DL] size head url=%s size=%d", url, size)
	return size
}
func DownloadAdditional(ctx *gin.Context) {
	downloadReq := model.AdditionalReq{}
	if err := json.Unmarshal([]byte(ctx.GetString("request_data")), &downloadReq); err != nil {
		panic(err)
	}
	log.Printf("[DL] additional req target_os=%s pkg_type=%d pkg_id=%d", downloadReq.TargetOs, downloadReq.PackageType, downloadReq.PackageID)
	pkgList := []model.AdditionalRes{}
	if SifCdnServer != "" {
		pkgType, pkgId := downloadReq.PackageType, downloadReq.PackageID
		var pkgInfo []PkgInfo
		err := MainEng.Table("download_m").Where("pkg_type = ? AND pkg_id = ? AND pkg_os = ?", pkgType, pkgId, downloadReq.TargetOs).
			Cols("pkg_id,pkg_order,pkg_size").
			OrderBy("pkg_id ASC, pkg_order ASC").Find(&pkgInfo)
		CheckErr(err)

		for _, pkg := range pkgInfo {
			url := fmt.Sprintf("%s/%s/archives/%d_%d_%d.zip", SifCdnServer, downloadReq.TargetOs, pkgType, pkg.Id, pkg.Order)
			size := pkgSizeOrRemote(pkg.Size, url)
			if size == 0 {
				log.Printf("[DL] missing url=%s", url)
				continue
			}
			pkgList = append(pkgList, model.AdditionalRes{
				Size: size,
				URL:  url,
			})
		}
	}

	urls := make([]string, 0, len(pkgList))
	for _, p := range pkgList {
		urls = append(urls, p.URL)
	}
	logSampleURLs("additional", urls, 5)

	addResp := model.AdditionalResp{
		ResponseData: pkgList,
		ReleaseInfo:  []any{},
		StatusCode:   200,
	}
	resp, err := json.Marshal(addResp)
	CheckErr(err)

	nonce := ctx.GetInt("nonce")
	nonce++

	ctx.Header("user_id", ctx.GetString("userid"))
	ctx.Header("authorize", fmt.Sprintf("consumerKey=lovelive_test&timeStamp=%d&version=1.1&token=%s&nonce=%d&user_id=%s&requestTimeStamp=%d", time.Now().Unix(), ctx.GetString("token"), nonce, ctx.GetString("userid"), ctx.GetInt64("req_time")))
	ctx.Header("X-Message-Sign", base64.StdEncoding.EncodeToString(encrypt.RSA_Sign_SHA1(resp, "privatekey.pem")))

	ctx.String(http.StatusOK, string(resp))
}

func DownloadBatch(ctx *gin.Context) {
	downloadReq := model.BatchReq{}
	if err := json.Unmarshal([]byte(ctx.GetString("request_data")), &downloadReq); err != nil {
		panic(err)
	}
	log.Printf("[DL] batch req os=%s pkg_type=%d client_version=%s excluded=%d", downloadReq.Os, downloadReq.PackageType, downloadReq.ClientVersion, len(downloadReq.ExcludedPackageIds))
	pkgList := []model.BatchRes{}
	if downloadReq.ClientVersion == config.PackageVersion && SifCdnServer != "" {
		pkgType := downloadReq.PackageType
		var pkgInfo []PkgInfo
		err := MainEng.Table("download_m").Where(builder.NotIn("pkg_id", downloadReq.ExcludedPackageIds)).Where("pkg_type = ? AND pkg_os = ?", pkgType, downloadReq.Os).
			Cols("pkg_id,pkg_order,pkg_size").
			OrderBy("pkg_id ASC, pkg_order ASC").Find(&pkgInfo)
		CheckErr(err)

		for _, pkg := range pkgInfo {
			url := fmt.Sprintf("%s/%s/archives/%d_%d_%d.zip", SifCdnServer, downloadReq.Os, pkgType, pkg.Id, pkg.Order)
			size := pkgSizeOrRemote(pkg.Size, url)
			if size == 0 {
				log.Printf("[DL] missing url=%s", url)
				continue
			}
			pkgList = append(pkgList, model.BatchRes{
				Size: size,
				URL:  url,
			})
		}
	}

	urls := make([]string, 0, len(pkgList))
	for _, p := range pkgList {
		urls = append(urls, p.URL)
	}
	logSampleURLs("batch", urls, 5)

	batchResp := model.BatchResp{
		ResponseData: pkgList,
		ReleaseInfo:  []any{},
		StatusCode:   200,
	}
	resp, err := json.Marshal(batchResp)
	CheckErr(err)

	nonce := ctx.GetInt("nonce")
	nonce++

	ctx.Header("user_id", ctx.GetString("userid"))
	ctx.Header("authorize", fmt.Sprintf("consumerKey=lovelive_test&timeStamp=%d&version=1.1&token=%s&nonce=%d&user_id=%s&requestTimeStamp=%d", time.Now().Unix(), ctx.GetString("token"), nonce, ctx.GetString("userid"), ctx.GetInt64("req_time")))
	ctx.Header("X-Message-Sign", base64.StdEncoding.EncodeToString(encrypt.RSA_Sign_SHA1(resp, "privatekey.pem")))

	ctx.String(http.StatusOK, string(resp))
}

func DownloadUpdate(ctx *gin.Context) {
	downloadReq := model.UpdateReq{}
	if err := json.Unmarshal([]byte(ctx.GetString("request_data")), &downloadReq); err != nil {
		panic(err)
	}
	log.Printf("[DL] update req target_os=%s external_version=%s", downloadReq.TargetOs, downloadReq.ExternalVersion)
	pkgList := []model.UpdateRes{}
	if SifCdnServer != "" {
		pkgType := 99
		ver := config.PackageVersion
		if downloadReq.ExternalVersion != "" {
			ver = downloadReq.ExternalVersion
		}
		var pkgInfo []PkgInfo
		err := MainEng.Table("download_m").Where("pkg_type = ? AND pkg_os = ?", pkgType, downloadReq.TargetOs).
			Cols("pkg_id,pkg_order,pkg_size").
			OrderBy("pkg_id ASC, pkg_order ASC").Find(&pkgInfo)
		CheckErr(err)

		for _, pkg := range pkgInfo {
			url := fmt.Sprintf("%s/%s/archives/%d_%d_%d.zip", SifCdnServer, downloadReq.TargetOs, pkgType, pkg.Id, pkg.Order)
			size := pkgSizeOrRemote(pkg.Size, url)
			if size == 0 {
				log.Printf("[DL] missing url=%s", url)
				continue
			}
			pkgList = append(pkgList, model.UpdateRes{
				Size:    size,
				URL:     url,
				Version: ver,
			})
		}

		patchFileUrl := fmt.Sprintf("%s/%s/archives/99_0_115.zip", SifCdnServer, downloadReq.TargetOs)
		size := pkgSizeOrRemote(0, patchFileUrl)
		if size > 0 {
			pkgList = append(pkgList, model.UpdateRes{
				Size:    size,
				URL:     patchFileUrl,
				Version: ver,
			})
		}
	}

	urls := make([]string, 0, len(pkgList))
	for _, p := range pkgList {
		urls = append(urls, p.URL)
	}
	logSampleURLs("update", urls, 5)

	updateResp := model.UpdateResp{
		ResponseData: pkgList,
		ReleaseInfo:  []any{},
		StatusCode:   200,
	}
	resp, err := json.Marshal(updateResp)
	CheckErr(err)

	// debug: write response to file for inspection
	if wd, err := os.Getwd(); err == nil {
		_ = os.WriteFile(filepath.Join(wd, "update_resp.json"), resp, 0644)
		if pretty, err := json.MarshalIndent(updateResp, "", "  "); err == nil {
			_ = os.WriteFile(filepath.Join(wd, "update_resp_debug.json"), pretty, 0644)
		}
		log.Printf("[DL] update_resp write dir=%s", wd)
	}

	nonce := ctx.GetInt("nonce")
	nonce++

	ctx.Header("user_id", ctx.GetString("userid"))
	ctx.Header("authorize", fmt.Sprintf("consumerKey=lovelive_test&timeStamp=%d&version=1.1&token=%s&nonce=%d&user_id=%s&requestTimeStamp=%d", time.Now().Unix(), ctx.GetString("token"), nonce, ctx.GetString("userid"), ctx.GetInt64("req_time")))
	ctx.Header("X-Message-Sign", base64.StdEncoding.EncodeToString(encrypt.RSA_Sign_SHA1(resp, "privatekey.pem")))

	ctx.String(http.StatusOK, string(resp))
}

func DownloadUrl(ctx *gin.Context) {
	// Extract SQL: SELECT CAST(pkg_type AS TEXT) || '_' || CAST(pkg_id AS TEXT) || '_' || CAST(pkg_order AS TEXT) || '.zip' AS zip_name FROM download_m ORDER BY pkg_type ASC,pkg_id ASC, pkg_order ASC;
	// Extract Cmd: cat list.txt | while read line; do; unzip -o $line; done
	downloadReq := model.UrlReq{}
	if err := json.Unmarshal([]byte(ctx.GetString("request_data")), &downloadReq); err != nil {
		panic(err)
	}
	urlList := []string{}
	for _, v := range downloadReq.PathList {
		urlList = append(urlList, fmt.Sprintf("%s/%s/extracted/%s", SifCdnServer, downloadReq.Os, strings.ReplaceAll(v, "\\", "")))
	}
	urlResp := model.UrlResp{
		ResponseData: model.UrlRes{
			UrlList: urlList,
		},
		ReleaseInfo: []any{},
		StatusCode:  200,
	}
	resp, err := json.Marshal(urlResp)
	CheckErr(err)

	nonce := ctx.GetInt("nonce")
	nonce++

	ctx.Header("user_id", ctx.GetString("userid"))
	ctx.Header("authorize", fmt.Sprintf("consumerKey=lovelive_test&timeStamp=%d&version=1.1&token=%s&nonce=%d&user_id=%s&requestTimeStamp=%d", time.Now().Unix(), ctx.GetString("token"), nonce, ctx.GetString("userid"), ctx.GetInt64("req_time")))
	ctx.Header("X-Message-Sign", base64.StdEncoding.EncodeToString(encrypt.RSA_Sign_SHA1(resp, "privatekey.pem")))

	ctx.String(http.StatusOK, string(resp))
}

func DownloadEvent(ctx *gin.Context) {
	eventResp := model.EventResp{
		ResponseData: []any{},
		ReleaseInfo:  []any{},
		StatusCode:   200,
	}
	resp, err := json.Marshal(eventResp)
	CheckErr(err)

	nonce := ctx.GetInt("nonce")
	nonce++

	ctx.Header("user_id", ctx.GetString("userid"))
	ctx.Header("authorize", fmt.Sprintf("consumerKey=lovelive_test&timeStamp=%d&version=1.1&token=%s&nonce=%d&user_id=%s&requestTimeStamp=%d", time.Now().Unix(), ctx.GetString("token"), nonce, ctx.GetString("userid"), ctx.GetInt64("req_time")))
	ctx.Header("X-Message-Sign", base64.StdEncoding.EncodeToString(encrypt.RSA_Sign_SHA1(resp, "privatekey.pem")))

	ctx.String(http.StatusOK, string(resp))
}
