package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/sha3"
)

type question struct {
	Text string `json:"text"`
	Inst bool   `json:"by_instructor"`
	By   string `json:"author_nic"`
}

type room struct {
	mx    sync.Mutex
	qs    []question
	icode string
	scode string
}

var rmx sync.Mutex
var rooms map[string]*room

func main() {
	rooms = make(map[string]*room)
	router := gin.Default()

	router.GET("/poll/:slug/:code", func(c *gin.Context) {
		rmx.Lock()
		defer rmx.Unlock()

		r, ok := rooms[c.Param("slug")]
		if !ok {
			c.String(http.StatusNotFound, "That room does not exist.")
			return
		}

		if c.Param("code") == r.scode {
			c.String(http.StatusForbidden, "This page is only available to instructors.")
			return
		}
		if c.Param("code") != r.icode {
			c.String(http.StatusUnauthorized, "Wrong code given for this room.")
			return
		}

		since, err := strconv.Atoi(c.DefaultQuery("since", "0"))
		if err != nil || since < 0 {
			c.String(http.StatusBadRequest, "Non-integer 'since' given")
			return
		}

		if since >= len(r.qs) {
			c.JSON(http.StatusNoContent, []question{})
			return
		}

		c.JSON(http.StatusOK, r.qs[since:])
	})

	router.GET("/room/:slug/:code", func(c *gin.Context) {
		rmx.Lock()
		defer rmx.Unlock()
		r, ok := rooms[c.Param("slug")]
		if !ok {
			h := sha3.Sum512([]byte(c.Param("code")))
			scode := hex.EncodeToString(h[:])[0:8]
			r = &room{
				qs: []question{{
					Text: fmt.Sprintf("Room created. Student code is '%s'", scode),
					Inst: true,
					By:   "<master>",
				}},
				icode: c.Param("code"),
				scode: scode,
			}
			rooms[c.Param("slug")] = r
		}

		if c.Param("code") != r.icode && c.Param("code") != r.scode {
			c.HTML(http.StatusUnauthorized, "bad.tmpl", gin.H{
				"room": c.Param("slug"),
			})
			return
		}

		revqs := make([]question, len(r.qs))
		for i := range r.qs {
			revqs[i] = r.qs[len(r.qs)-i-1]
		}
		c.HTML(http.StatusOK, "room.tmpl", gin.H{
			"room":       c.Param("slug"),
			"instructor": c.Param("code") == r.icode,
			"submitted":  c.DefaultQuery("submitted", "0"),
			"qs":         revqs,
			"scode":      r.scode,
		})
	})

	router.POST("/room/:slug/:code", func(c *gin.Context) {
		rmx.Lock()
		defer rmx.Unlock()

		r, ok := rooms[c.Param("slug")]
		if !ok {
			c.String(http.StatusNotFound, "That room does not exist.")
			return
		}

		if c.Param("code") != r.icode && c.Param("code") != r.scode {
			c.String(http.StatusUnauthorized, "Wrong code given for this room.")
			return
		}

		q := c.PostForm("question")
		if strings.TrimSpace(q) == "" {
			c.String(http.StatusBadRequest, "Question was empty.")
			return
		}

		nick := strings.TrimSpace(c.DefaultPostForm("nick", ""))
		r.qs = append(r.qs, question{
			Text: q,
			Inst: c.Param("code") == r.icode,
			By:   nick,
		})
		to := *c.Request.URL
		to.RawQuery = "submitted=1"
		c.Redirect(http.StatusFound, to.RequestURI())
	})

	router.Static("/static", "./static")
	router.LoadHTMLGlob("./templates/*")

	// Listen and server on 0.0.0.0:8080
	router.Run(":8080")
}
