package custd

import "context"

var appMessageSurfaceRows = []surfaceRow{
	{id: "admin-api GET /api/v1/app/messages/session", path: "/api/v1/app/messages/session", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.Session(context.Background())
		return err
	}},
	{id: "admin-api POST /api/v1/app/messages/batch", path: "/api/v1/app/messages/batch", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.Import(context.Background(), &AppMessageBatchRequest{Source: AppMessageSourceSMS})
		return err
	}},
	{id: "admin-api POST /api/v1/app/messages/search", path: "/api/v1/app/messages/search", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.Search(context.Background(), &AppMessageSearchRequest{Query: "test"})
		return err
	}},
	{id: "admin-api GET /api/v1/app/messages/{messageUuid}", path: "/api/v1/app/messages/" + appArchiveTestUUID, call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.Get(context.Background(), appArchiveTestUUID)
		return err
	}},
	{id: "admin-api DELETE /api/v1/app/messages/sources/{source}", path: "/api/v1/app/messages/sources/sms", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.DeleteSource(context.Background(), AppMessageSourceSMS)
		return err
	}},
	{id: "admin-api PUT /api/v1/app/messages/{messageUuid}/media", path: "/api/v1/app/messages/" + appArchiveTestUUID + "/media", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.UploadMedia(context.Background(), &AppMessageMediaUpload{
			MessageUUID: appArchiveTestUUID, Image: []byte{0xff, 0xd8, 0xff, 0xd9},
		})
		return err
	}},
	{id: "admin-api GET /api/v1/app/messages/{messageUuid}/media/{mediaUuid}", path: "/api/v1/app/messages/" + appArchiveTestUUID + "/media/" + appArchiveTestUUID, call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.DownloadMedia(context.Background(), AppMessageMediaRef{
			MessageUUID: appArchiveTestUUID, MediaUUID: appArchiveTestUUID,
		})
		return err
	}},

	{id: "admin-api GET /api/v1/app/messages/connector", path: "/api/v1/app/messages/connector", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.GetConnector(context.Background())
		return err
	}},
	{id: "admin-api POST /api/v1/app/messages/connector/control", path: "/api/v1/app/messages/connector/control", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.ControlConnector(context.Background(), AppMessageConnectorControl{Action: AppMessageConnectorPause})
		return err
	}},
	{id: "admin-api POST /api/v1/app/messages/connector/enroll", path: "/api/v1/app/messages/connector/enroll", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.EnrollConnector(context.Background())
		return err
	}},
	{id: "admin-api POST /api/v1/app/messages/connector/report", path: "/api/v1/app/messages/connector/report", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.ReportConnector(context.Background(), AppMessageConnectorReport{ConnectionState: AppMessageConnectorSyncing})
		return err
	}},
	{id: "admin-api GET /api/v1/app/messages/connector/session", path: "/api/v1/app/messages/connector/session", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.ConnectorSession(context.Background())
		return err
	}},
	{id: "admin-api POST /api/v1/app/messages/mutations", path: "/api/v1/app/messages/mutations", call: func(e *surfaceEnv) error {
		_, err := e.client.AppMessages.Mutate(context.Background(), AppMessageMutationBatch{Source: AppMessageSourceWhatsApp, SourceGeneration: 2, Mutations: []AppMessageMutation{}})
		return err
	}},
}
