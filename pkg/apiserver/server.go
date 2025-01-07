package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/emicklei/go-restful"
)

const (
	SERVICE_VERSION = "v1-alpha"
)

func RegisterService() *restful.WebService {
	service := new(restful.WebService)

	service.Path("/" + SERVICE_VERSION).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	service.Route(service.GET("/metrics").To(getMetrics))

	return service

}

func StartWebServer(host string, port int) {
	log.Printf("Starting server on %d, endpoint is %s", port, "/"+SERVICE_VERSION)
	restful.Add(RegisterService())
	log.Fatal(http.ListenAndServe(host+":"+strconv.Itoa(port), nil))
}

type metricsVO struct {
	GitLink          string     `json:"link"`
	Ecosystems       *string    `json:"ecosystems"`
	CreatedSince     *time.Time `json:"createdSince"`
	UpdatedSince     *time.Time `json:"updatedSince"`
	ContributorCount *int       `json:"contributorCount"`
	OrgCount         *int       `json:"orgCount"`
	CommitFrequency  *float64   `json:"commitFrequency"`
	DepsDevCount     *int       `json:"depsDevCount"`
	DepsDistroScore  *float64   `json:"depsDistroScore"`
	License          *string    `json:"license"`
	Language         *string    `json:"language"`
	Industry         *string    `json:"industry"`
	Domestic         *bool      `json:"domestic"`
	Score            *float64   `json:"score"`
	// Rank             int       `json:"rank"`
}

const MAX_ALLOWED_TAKE = 10000

func getMetrics(request *restful.Request, response *restful.Response) {

	startStr := request.QueryParameter("start")
	if startStr == "" {
		startStr = "0"
	}
	start, err := strconv.Atoi(startStr)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, "Invalid start parameter")
		return
	}

	takeStr := request.QueryParameter("take")
	if takeStr == "" {
		takeStr = "100"
	}
	take, err := strconv.Atoi(takeStr)

	conn, err := storage.GetDefaultDatabaseConnection()
	defer conn.Close()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	var total int

	if take > MAX_ALLOWED_TAKE {
		response.WriteErrorString(http.StatusBadRequest, "take parameter is too large")
	}

	r := conn.QueryRow(`SELECT COUNT(*) FROM git_metrics_prod WHERE scores IS NOT NULL`)
	if r == nil {
		response.WriteErrorString(http.StatusInternalServerError, "No data found")
		return
	}
	r.Scan(&total)

	rows, err := conn.Query(`SELECT
		gm.git_link AS git_link,
		ecosystem,
		created_since,
		updated_since,
		contributor_count,
		org_count,
		commit_frequency,
		depsdev_count,
		deps_distro,
		license,
		language,
		CASE WHEN industry IS NULL
			THEN 'unknown'
			WHEN industry = 0
			THEN 'test0'
			ELSE 'other'
		END AS industry,
		gr.domestic AS domestic,
		scores
	FROM git_metrics_prod gm
	LEFT JOIN git_repositories gr ON gm.git_link = gr.git_link
	WHERE scores IS NOT NULL
	ORDER BY scores DESC
	OFFSET $1 LIMIT $2`, start, take)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "Fetch data error")
		log.Print(err)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("X-From", "criticality_score")
	response.WriteHeader(200)
	response.Write([]byte(`{"total":` + strconv.Itoa(total) + `,"data":[`))

	var first = true

	for rows.Next() {
		if !first {
			response.Write([]byte(","))
		}
		first = false

		var metrics metricsVO
		err = rows.Scan(
			&metrics.GitLink,
			&metrics.Ecosystems,
			&metrics.CreatedSince,
			&metrics.UpdatedSince,
			&metrics.ContributorCount,
			&metrics.OrgCount,
			&metrics.CommitFrequency,
			&metrics.DepsDevCount,
			&metrics.DepsDistroScore,
			&metrics.License,
			&metrics.Language,
			&metrics.Industry,
			&metrics.Domestic,
			&metrics.Score)
		if err != nil {
			log.Print("err at scan: ", err)
			return
		}
		jsonBytes, err := json.Marshal(metrics)
		if err != nil {
			log.Print("err at json marshal: ", err)
			return
		}
		response.Write(jsonBytes)
		response.Flush()
	}
	response.Write([]byte("]}"))
	response.Flush()
}
