package edit

import (
	"errors"
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func profileEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type ProfileRequest struct {
		Name string `json:"name"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{

			http.MethodPost: endpoint.JsonBodyMethod(func(request endpoint.Request[ProfileRequest]) error {
				cleanName := strings.TrimSpace(request.Body.Name)

				if cleanName == "" {
					return errors.New("profile name can not be empty")
				}

				graphInstance.SaveProfile(cleanName)

				saver.Save()
				return nil
			}),

			http.MethodDelete: endpoint.JsonBodyMethod(func(request endpoint.Request[ProfileRequest]) error {
				err := graphInstance.DeleteProfile(request.Body.Name)
				if err != nil {
					return err
				}
				saver.Save()

				return nil
			}),
		},
	}
}

func applyProfileEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type ApplyProfileRequest struct {
		Name string `json:"name"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonBodyMethod(func(request endpoint.Request[ApplyProfileRequest]) error {
				err := graphInstance.LoadProfile(request.Body.Name)
				if err != nil {
					return err
				}
				saver.Save()
				return nil
			}),
		},
	}
}

func renameProfileEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type RenameProfileRequest struct {
		Original string `json:"original"`
		New      string `json:"new"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonBodyMethod(func(request endpoint.Request[RenameProfileRequest]) error {
				err := graphInstance.RenameProfile(request.Body.Original, request.Body.New)
				if err != nil {
					return err
				}
				saver.Save()
				return nil
			}),
		},
	}
}

func overwriteProfileEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type OverwriteProfileRequest struct {
		Name string `json:"name"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonBodyMethod(func(request endpoint.Request[OverwriteProfileRequest]) error {
				graphInstance.SaveProfile(request.Body.Name)
				saver.Save()
				return nil
			}),
		},
	}
}
