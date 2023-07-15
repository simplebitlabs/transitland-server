package authz

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/internal/generated/azpb"
	"github.com/interline-io/transitland-server/internal/util"
	"github.com/interline-io/transitland-server/model"
)

func NewServer(checker model.Checker) (http.Handler, error) {
	router := chi.NewRouter()

	/////////////////
	// USERS
	/////////////////

	router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.UserList(r.Context(), &azpb.UserListRequest{Q: r.URL.Query().Get("q")})
		handleJson(w, ret, err)
	})
	router.Get("/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.User(r.Context(), &azpb.UserRequest{Id: chi.URLParam(r, "user_id")})
		handleJson(w, ret, err)
	})

	/////////////////
	// TENANTS
	/////////////////

	router.Get("/tenants", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.TenantList(r.Context(), &azpb.TenantListRequest{})
		handleJson(w, ret, err)
	})
	router.Get("/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.TenantPermissions(r.Context(), &azpb.TenantRequest{Id: checkId(r, "tenant_id")})
		handleJson(w, ret, err)
	})
	router.Post("/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		check := azpb.Tenant{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		check.Id = checkId(r, "tenant_id")
		_, err := checker.TenantSave(r.Context(), &azpb.TenantSaveRequest{Tenant: &check})
		handleJson(w, nil, err)
	})
	router.Post("/tenants/{tenant_id}/groups", func(w http.ResponseWriter, r *http.Request) {
		check := azpb.Group{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err := checker.TenantCreateGroup(r.Context(), &azpb.TenantCreateGroupRequest{Id: checkId(r, "tenant_id"), Group: &check})
		handleJson(w, nil, err)
	})
	router.Post("/tenants/{tenant_id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		check := &azpb.EntityRelation{}
		if err := parseJson(r.Body, check); err != nil {
			handleJson(w, nil, err)
		}
		entId := checkId(r, "tenant_id")
		_, err := checker.TenantAddPermission(r.Context(), &azpb.TenantModifyPermissionRequest{Id: entId, EntityRelation: check})
		handleJson(w, nil, err)
	})
	router.Delete("/tenants/{tenant_id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		check := &azpb.EntityRelation{}
		if err := parseJson(r.Body, check); err != nil {
			handleJson(w, nil, err)
		}
		entId := checkId(r, "tenant_id")
		_, err := checker.TenantRemovePermission(r.Context(), &azpb.TenantModifyPermissionRequest{Id: entId, EntityRelation: check})
		handleJson(w, nil, err)
	})

	/////////////////
	// GROUPS
	/////////////////

	router.Get("/groups", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.GroupList(r.Context(), &azpb.GroupListRequest{})
		handleJson(w, ret, err)
	})
	router.Post("/groups/{group_id}", func(w http.ResponseWriter, r *http.Request) {
		check := azpb.Group{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		check.Id = checkId(r, "group_id")
		_, err := checker.GroupSave(r.Context(), &azpb.GroupSaveRequest{Group: &check})
		handleJson(w, nil, err)
	})
	router.Get("/groups/{group_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.GroupPermissions(r.Context(), &azpb.GroupRequest{Id: checkId(r, "group_id")})
		handleJson(w, ret, err)
	})
	router.Post("/groups/{group_id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		check := &azpb.EntityRelation{}
		if err := parseJson(r.Body, check); err != nil {
			handleJson(w, nil, err)
		}
		entId := checkId(r, "group_id")
		_, err := checker.GroupAddPermission(r.Context(), &azpb.GroupModifyPermissionRequest{Id: entId, EntityRelation: check})
		handleJson(w, nil, err)
	})
	router.Delete("/groups/{group_id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		check := &azpb.EntityRelation{}
		if err := parseJson(r.Body, check); err != nil {
			handleJson(w, nil, err)
		}
		entId := checkId(r, "group_id")
		_, err := checker.GroupRemovePermission(r.Context(), &azpb.GroupModifyPermissionRequest{Id: entId, EntityRelation: check})
		handleJson(w, nil, err)
	})
	router.Post("/groups/{group_id}/tenant", func(w http.ResponseWriter, r *http.Request) {
		check := azpb.GroupSetTenantRequest{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		check.Id = checkId(r, "group_id")
		_, err := checker.GroupSetTenant(r.Context(), &check)
		handleJson(w, nil, err)
	})

	/////////////////
	// FEEDS
	/////////////////

	router.Get("/feeds", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedList(r.Context(), &azpb.FeedListRequest{})
		handleJson(w, ret, err)
	})
	router.Get("/feeds/{feed_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedPermissions(r.Context(), &azpb.FeedRequest{Id: checkId(r, "feed_id")})
		handleJson(w, ret, err)
	})
	router.Post("/feeds/{feed_id}/group", func(w http.ResponseWriter, r *http.Request) {
		check := azpb.FeedSetGroupRequest{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		check.Id = checkId(r, "feed_id")
		_, err := checker.FeedSetGroup(r.Context(), &check)
		handleJson(w, nil, err)
	})

	/////////////////
	// FEED VERSIONS
	/////////////////

	router.Get("/feed_versions", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedVersionList(r.Context(), &azpb.FeedVersionListRequest{})
		handleJson(w, ret, err)
	})
	router.Get("/feed_versions/{feed_version_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedVersionPermissions(r.Context(), &azpb.FeedVersionRequest{Id: checkId(r, "feed_version_id")})
		handleJson(w, ret, err)
	})
	router.Post("/feed_versions/{feed_version_id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		check := &azpb.EntityRelation{}
		if err := parseJson(r.Body, check); err != nil {
			handleJson(w, nil, err)
		}
		entId := checkId(r, "feed_version_id")
		_, err := checker.FeedVersionAddPermission(r.Context(), &azpb.FeedVersionModifyPermissionRequest{Id: entId, EntityRelation: check})
		handleJson(w, nil, err)
	})
	router.Delete("/feed_versions/{feed_version_id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		check := &azpb.EntityRelation{}
		if err := parseJson(r.Body, check); err != nil {
			handleJson(w, nil, err)
		}
		entId := checkId(r, "feed_version_id")
		_, err := checker.FeedVersionRemovePermission(r.Context(), &azpb.FeedVersionModifyPermissionRequest{Id: entId, EntityRelation: check})
		handleJson(w, nil, err)
	})

	return router, nil
}

func handleJson(w http.ResponseWriter, ret any, err error) {
	if err == ErrUnauthorized {
		log.Error().Err(err).Msg("unauthorized")
		http.Error(w, util.MakeJsonError(http.StatusText(http.StatusUnauthorized)), http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Error().Err(err).Msg("admin api error")
		http.Error(w, util.MakeJsonError(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}
	if ret == nil {
		ret = map[string]bool{"success": true}
	}
	jj, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

func checkId(r *http.Request, key string) int64 {
	v, _ := strconv.Atoi(chi.URLParam(r, key))
	return int64(v)
}

func parseJson(r io.Reader, v any) error {
	data, err := ioutil.ReadAll(io.LimitReader(r, 1_000_000))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}