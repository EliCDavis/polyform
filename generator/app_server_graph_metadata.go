package generator

// func graphMetadataEndpoint(as *AppServer) endpoint.Handler {

// 	type EditRequest map[string]any

// 	type EmptyResponse struct{}

// 	return endpoint.Handler{
// 		Methods: map[string]endpoint.Method{
// 			http.MethodPost: endpoint.JsonMethod(
// 				func(request endpoint.Request[EditRequest]) (EmptyResponse, error) {

// 					// We're making the assumption the url starts like this,
// 					// so assert it.
// 					if strings.Index(request.Url, "/graph/metadata") != 0 {
// 						panic("url should begin with /graph/metadata")
// 					}

// 					metadataPath := request.Url[len("/graph/metadata"):]

// 					if metadataPath[0] == '/' {
// 						metadataPath = metadataPath[1:]
// 					}

// 					if len(metadataPath) > 0 {
// 						metadataPath = "." + strings.Replace(metadataPath, "/", ".", -1)
// 					}

// 					as.app.nodeMetadata.Set(request.Body.NodeID+metadataPath, request.Body)
// 					as.AutosaveGraph()
// 					return EmptyResponse{}, nil
// 				},
// 			),
// 		},
// 	}
// }
