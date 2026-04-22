package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"honoka-chan/encrypt"
	"honoka-chan/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)


func writeSigned(ctx *gin.Context, resp []byte) {
	nonce := ctx.GetInt("nonce")
	nonce++

	ctx.Header("user_id", ctx.GetString("userid"))
	ctx.Header("authorize", fmt.Sprintf("consumerKey=lovelive_test&timeStamp=%d&version=1.1&token=%s&nonce=%d&user_id=%s&requestTimeStamp=%d", time.Now().Unix(), ctx.GetString("token"), nonce, ctx.GetString("userid"), ctx.GetInt64("req_time")))
	ctx.Header("X-Message-Sign", base64.StdEncoding.EncodeToString(encrypt.RSA_Sign_SHA1(resp, "privatekey.pem")))

	ctx.String(http.StatusOK, string(resp))
}

func writeEmptySetDeck(ctx *gin.Context) {
	resp, _ := json.Marshal(model.SetDeckResp{
		ResponseData: []any{},
		ReleaseInfo:  []any{},
		StatusCode:   200,
	})
	writeSigned(ctx, resp)
}

func writeEmptyAwardSet(ctx *gin.Context) {
	resp, _ := json.Marshal(model.AwardSetResp{
		ResponseData: []any{},
		ReleaseInfo:  []any{},
		StatusCode:   200,
	})
	writeSigned(ctx, resp)
}
func SetDisplayRank(ctx *gin.Context) {
	dispResp := model.SetDisplayRankResp{
		ResponseData: []any{},
		ReleaseInfo:  []any{},
		StatusCode:   200,
	}
	resp, err := json.Marshal(dispResp)
	CheckErr(err)

	nonce := ctx.GetInt("nonce")
	nonce++

	ctx.Header("user_id", ctx.GetString("userid"))
	ctx.Header("authorize", fmt.Sprintf("consumerKey=lovelive_test&timeStamp=%d&version=1.1&token=%s&nonce=%d&user_id=%s&requestTimeStamp=%d", time.Now().Unix(), ctx.GetString("token"), nonce, ctx.GetString("userid"), ctx.GetInt64("req_time")))
	ctx.Header("X-Message-Sign", base64.StdEncoding.EncodeToString(encrypt.RSA_Sign_SHA1(resp, "privatekey.pem")))

	ctx.String(http.StatusOK, string(resp))
}

func SetDeck(ctx *gin.Context) {
	userId, err := strconv.Atoi(ctx.GetString("userid"))
	if err != nil {
		writeEmptySetDeck(ctx)
		return
	}

	deckReq := model.UnitDeckReq{}
	if err := json.Unmarshal([]byte(ctx.PostForm("request_data")), &deckReq); err != nil {
		writeEmptySetDeck(ctx)
		return
	}

	// DemoUserDemoUserDemoUser?
	// UserEng.ShowSQL(true)
	session := UserEng.NewSession()
	defer session.Close()
	fail := func() {
		_ = session.Rollback()
		writeEmptySetDeck(ctx)
	}
	if err := session.Begin(); err != nil {
		fail()
		return
	}

	// DemoUserDemoUserDemoUserDemoUser?
	var userDeckId []int
	err = session.Table("user_deck_m").Cols("id").Where("user_id = ?", userId).Find(&userDeckId)
	if err != nil {
		fail()
		return
	}

	// DemoUserDemoUserDemoUserDemoUserDemoUserDemoUserDemoUser?
	_, err = session.Table("deck_unit_m").In("user_deck_id", userDeckId).Delete()
	if err != nil {
		fail()
		return
	}

	// DemoUserDemoUserDemoUserDemoUserDemoUserDemoUser
	_, err = session.Table("user_deck_m").In("id", userDeckId).Delete()
	if err != nil {
		fail()
		return
	}

	// DemoUserDemoUserDemoUserDemoUser
	for _, deck := range deckReq.UnitDeckList {
		// DemoUserDemoUserDemoUserDemoUser
		userDeck := model.UserDeckData{
			DeckID:     deck.UnitDeckID,
			MainFlag:   deck.MainFlag,
			DeckName:   deck.DeckName,
			UserID:     userId,
			InsertDate: time.Now().Unix(),
		}
		_, err = session.Table("user_deck_m").Insert(&userDeck)
		if err != nil {
			fail()
			return
		}
		userDeckId := userDeck.ID
		// fmt.Println("DemoUserDemoUser?ID:", userDeckId)

		// DemoUserDemoUserDemoUserDemoUser?
		for _, unit := range deck.UnitDeckDetail {
			// DemoUserDemoUserDemoUser
			newUnitData := model.UnitData{}
			exists, err := session.Table("user_unit_m").Where("unit_owning_user_id = ?", unit.UnitOwningUserID).Exist()
			if err != nil {
				fail()
				return
			}
			if exists {
				// fmt.Println("DemoUserDemoUserDemoUserDemoUserDemoUserDemoUserDemoUser?")
				_, err = session.Table("user_unit_m").Where("unit_owning_user_id = ?", unit.UnitOwningUserID).Get(&newUnitData)
				if err != nil {
					fail()
					return
				}
			} else {
				exists, err = MainEng.Table("common_unit_m").Where("unit_owning_user_id = ?", unit.UnitOwningUserID).Exist()
				if err != nil {
					fail()
					return
				}
				if exists {
					// fmt.Println("DemoUserDemoUserDemoUserDemoUserDemoUserDemoUser")
					_, err = MainEng.Table("common_unit_m").Where("unit_owning_user_id = ?", unit.UnitOwningUserID).Get(&newUnitData)
					if err != nil {
						fail()
						return
					}
				} else {
					// fmt.Println("DemoUserDemoUserDemoUserDemoUser?")
					fail()
					return
				}
			}
			// fmt.Println("DemoUserDemoUserDemoUserDemoUser?:", newUnitData)

			// DemoUserDemoUserDemoUserDemoUserDemoUserDemoUser?
			newUnitDeckData := model.UnitDeckData{}
			b, err := json.Marshal(newUnitData)
			if err != nil {
				fail()
				return
			}
			if err = json.Unmarshal(b, &newUnitDeckData); err != nil {
				fail()
				return
			}
			newUnitDeckData.BeforeLove = newUnitDeckData.MaxLove
			newUnitDeckData.Position = unit.Position
			newUnitDeckData.UserDeckID = userDeckId
			newUnitDeckData.InsertData = time.Now().Unix()

			_, err = session.Table("deck_unit_m").Insert(&newUnitDeckData)
			if err != nil {
				fail()
				return
			}
		}
	}

	// DemoUserDemoUserDemoUser
	if err = session.Commit(); err != nil {
		fail()
		return
	}

	writeEmptySetDeck(ctx)
}
func SetDeckName(ctx *gin.Context) {
	userId, err := strconv.Atoi(ctx.GetString("userid"))
	if err != nil {
		writeEmptySetDeck(ctx)
		return
	}

	deckReq := model.DeckNameReq{}
	if err := json.Unmarshal([]byte(ctx.PostForm("request_data")), &deckReq); err != nil {
		writeEmptySetDeck(ctx)
		return
	}

	exists, err := UserEng.Table("user_deck_m").Where("user_id = ? AND deck_id = ?", userId, deckReq.UnitDeckID).Exist()
	if err != nil || !exists {
		writeEmptySetDeck(ctx)
		return
	}
	userDeck := model.UserDeckData{
		DeckName: deckReq.DeckName,
	}
	_, err = UserEng.Table("user_deck_m").Update(&userDeck, &model.UserDeckData{
		UserID: userId,
		DeckID: deckReq.UnitDeckID,
	})
	if err != nil {
		writeEmptySetDeck(ctx)
		return
	}

	writeEmptySetDeck(ctx)
}
func WearAccessory(ctx *gin.Context) {
	fmt.Println(ctx.PostForm("request_data"))
	req := model.WearAccessoryReq{}
	if err := json.Unmarshal([]byte(ctx.PostForm("request_data")), &req); err != nil {
		writeEmptyAwardSet(ctx)
		return
	}

	// UserEng.ShowSQL(true)
	// DemoUserDemoUserDemoUser?
	session := UserEng.NewSession()
	defer session.Close()
	fail := func() {
		_ = session.Rollback()
		writeEmptyAwardSet(ctx)
	}
	if err := session.Begin(); err != nil {
		fail()
		return
	}

	// DemoUserDemoUserDemoUser
	for _, v := range req.Remove {
		fmt.Println("Remove:", v.AccessoryOwningUserID, v.UnitOwningUserID)
		_, err := session.Table("accessory_wear_m").
			Where("accessory_owning_user_id = ? AND unit_owning_user_id = ? AND user_id = ?", v.AccessoryOwningUserID, v.UnitOwningUserID, ctx.GetString("userid")).
			Delete()
		if err != nil {
			fail()
			return
		}
	}

	// DemoUserDemoUserDemoUser
	for _, v := range req.Wear {
		fmt.Println("Wear:", v.AccessoryOwningUserID, v.UnitOwningUserID)
		data := model.AccessoryWearData{
			AccessoryOwningUserID: v.AccessoryOwningUserID,
			UnitOwningUserID:      v.UnitOwningUserID,
			UserId:                ctx.GetString("userid"),
		}
		_, err := session.Table("accessory_wear_m").Insert(&data)
		if err != nil {
			fail()
			return
		}
	}

	// DemoUserDemoUserDemoUser
	if err := session.Commit(); err != nil {
		fail()
		return
	}

	writeEmptyAwardSet(ctx)
}
func RemoveSkillEquip(ctx *gin.Context) {
	fmt.Println(ctx.PostForm("request_data"))
	req := model.SkillEquipReq{}
	if err := json.Unmarshal([]byte(ctx.PostForm("request_data")), &req); err != nil {
		writeEmptyAwardSet(ctx)
		return
	}

	// UserEng.ShowSQL(true)
	// DemoUserDemoUserDemoUser?
	session := UserEng.NewSession()
	defer session.Close()
	fail := func() {
		_ = session.Rollback()
		writeEmptyAwardSet(ctx)
	}
	if err := session.Begin(); err != nil {
		fail()
		return
	}

	// DemoUserDemoUserDemoUser
	for _, v := range req.Remove {
		fmt.Println("Remove:", v.UnitOwningUserID, v.UnitRemovableSkillID)
		_, err := session.Table("skill_equip_m").
			Where("unit_removable_skill_id = ? AND unit_owning_user_id = ? AND user_id = ?", v.UnitRemovableSkillID, v.UnitOwningUserID, ctx.GetString("userid")).
			Delete()
		if err != nil {
			fail()
			return
		}
	}

	// DemoUserDemoUserDemoUser
	for _, v := range req.Equip {
		fmt.Println("Equip:", v.UnitOwningUserID, v.UnitRemovableSkillID)
		data := model.SkillEquipData{
			UnitRemovableSkillId: v.UnitRemovableSkillID,
			UnitOwningUserID:     v.UnitOwningUserID,
			UserId:               ctx.GetString("userid"),
		}
		_, err := session.Table("skill_equip_m").Insert(&data)
		if err != nil {
			fail()
			return
		}
	}

	// DemoUserDemoUserDemoUser
	if err := session.Commit(); err != nil {
		fail()
		return
	}

	writeEmptyAwardSet(ctx)
}
