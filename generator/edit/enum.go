package edit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
)

const enumInstanceEndpointPath = "/enum/instance/"

type EnumBody struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Members     []enum.Member     `json:"members"`
	Renamed     map[string]string `json:"renamed,omitempty"`
}

type EnumResponse struct {
	ID              string          `json:"id"`
	Type            string          `json:"type"`
	NodeType        schema.NodeType `json:"nodeType"`
	MembersNodeType schema.NodeType `json:"membersNodeType"`
}

func enumInstanceEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	respond := func(id string) (EnumResponse, error) {
		nodeType, err := graphInstance.EnumParameterNodeType(id)
		if err != nil {
			return EnumResponse{}, err
		}
		membersType, err := graphInstance.EnumMembersNodeType(id)
		if err != nil {
			return EnumResponse{}, err
		}
		return EnumResponse{ID: id, Type: enum.RuntimeTypePath(id), NodeType: nodeType, MembersNodeType: membersType}, nil
	}

	idOf := func(url string) (string, error) {
		id := strings.TrimPrefix(url, enumInstanceEndpointPath)
		if id == "" {
			return "", fmt.Errorf("enum id is required")
		}
		return id, nil
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(func(request endpoint.Request[EnumBody]) (EnumResponse, error) {
				id, err := graphInstance.CreateEnum(request.Body.Name, request.Body.Description, request.Body.Members)
				if err != nil {
					return EnumResponse{}, err
				}
				saver.Save()
				return respond(id)
			}),

			http.MethodPut: endpoint.JsonMethod(func(request endpoint.Request[EnumBody]) (EnumResponse, error) {
				id, err := idOf(request.Url)
				if err != nil {
					return EnumResponse{}, err
				}
				err = graphInstance.UpdateEnum(id, graph.EnumUpdate{
					Name:        request.Body.Name,
					Description: request.Body.Description,
					Members:     request.Body.Members,
					Renamed:     request.Body.Renamed,
				})
				if err != nil {
					return EnumResponse{}, err
				}
				saver.Save()
				return respond(id)
			}),

			http.MethodDelete: endpoint.Func(func(request *http.Request) error {
				id, err := idOf(request.URL.Path)
				if err != nil {
					return err
				}
				if err := graphInstance.DeleteEnum(id); err != nil {
					return err
				}
				saver.Save()
				return nil
			}),
		},
	}
}
