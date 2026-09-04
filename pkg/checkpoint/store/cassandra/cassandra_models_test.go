package cassandra

import (
	"reflect"
	"testing"

	"github.com/SneaksAndData/nexus-core/pkg/util_test"
	"k8s.io/apimachinery/pkg/util/diff"
)

func TestCheckpointedRequest_ToCqlModel(t *testing.T) {
	fakeRequest := util_test.GetFakeRequest(false)
	cqlModel, err := ToCassandraModel(fakeRequest)
	if err != nil {
		t.Errorf("Error when creating CQL model: %s", err)
	}
	expectedCqlModel := &CheckpointCassandraModel{
		Algorithm:               fakeRequest.Algorithm,
		Id:                      fakeRequest.Id,
		LifecycleStage:          "RUNNING",
		PayloadUri:              fakeRequest.PayloadUri,
		ResultUri:               fakeRequest.ResultUri,
		AlgorithmFailureCause:   fakeRequest.AlgorithmFailureCause,
		AlgorithmFailureDetails: fakeRequest.AlgorithmFailureDetails,
		ReceivedByHost:          fakeRequest.ReceivedByHost,
		ReceivedAt:              fakeRequest.ReceivedAt,
		SentAt:                  fakeRequest.SentAt,
		AppliedConfiguration:    "b64__eyJjb250YWluZXIiOnsiaW1hZ2UiOiJ0ZXN0LmlvIiwicmVnaXN0cnkiOiJhbGdvcml0aG1zL3Rlc3QiLCJ2ZXJzaW9uVGFnIjoidjEuMC4wIiwic2VydmljZUFjY291bnROYW1lIjoidGVzdC1zYSJ9LCJjb21wdXRlUmVzb3VyY2VzIjp7ImNwdUxpbWl0IjoiMTAwMG0iLCJtZW1vcnlMaW1pdCI6IjIwMDBNaSIsImRlZmF1bHRSZXNvdXJjZVF1b3RhIjoiIiwibGltaXRzIjpudWxsfSwid29ya2dyb3VwUmVmIjp7Im5hbWUiOiJ0ZXN0LXdvcmtncm91cCIsImdyb3VwIjoibmV4dXMtd29ya2dyb3VwLmlvIiwia2luZCI6IkthcnBlbnRlcldvcmtncm91cFYxIn0sImNvbW1hbmQiOiJweXRob24iLCJhcmdzIjpbImpvYi5weSIsIi0tc2FzLXVyaT0lcyIsIi0tcmVxdWVzdC1pZD0lcyIsIi0tYXJnMT10cnVlIl0sInBheWxvYWRDb25maWd1cmF0aW9uIjp7InBheWxvYWRWYWxpZEZvciI6IjI0aCIsInBheWxvYWRTZXJpYWxpemF0aW9uTW9kZSI6ImJhY2tlbmQifSwicnVudGltZUVudmlyb25tZW50Ijp7ImRlYWRsaW5lU2Vjb25kcyI6MTIwLCJtYXhpbXVtUmV0cmllcyI6M30sImRhdGFkb2dJbnRlZ3JhdGlvblNldHRpbmdzIjp7Im1vdW50RGF0YWRvZ1NvY2tldCI6dHJ1ZX19",
		ConfigurationOverrides:  "b64__e30=",
		ContentHash:             fakeRequest.ContentHash,
		LastModified:            fakeRequest.LastModified,
		Tag:                     fakeRequest.Tag,
		ApiVersion:              fakeRequest.ApiVersion,
		JobUid:                  fakeRequest.JobUid,
		Parent:                  "b64__e30=",
	}

	if !reflect.DeepEqual(expectedCqlModel, cqlModel) {
		t.Fatalf("Failed to convert request to a Cassandra model %s: values do not match", diff.ObjectGoPrintSideBySide(expectedCqlModel, cqlModel))
	}
	t.Log("cassandra.ToCassandraModel() returns correct result")
}

func TestCheckpointedRequest_FromCqlModel(t *testing.T) {
	fakeRequest := util_test.GetFakeRequest(false)
	cqlModel, err := ToCassandraModel(fakeRequest)
	if err != nil {
		t.Fatalf("Error when creating CQL model: %s", err)
	}
	fakeRequestFromModel, err := cqlModel.FromCassandraModel()

	if err != nil {
		t.Fatalf("Error when converting a Cassandra model back to a checkpoint: %s", err)
	}

	if !reflect.DeepEqual(fakeRequest, fakeRequestFromModel) {
		t.Fatalf("Failed to deserialize a checkpoint from its cql model %s: values do not match", diff.ObjectGoPrintSideBySide(fakeRequest, fakeRequestFromModel))
	}
	t.Log("cassandra.ToCassandraModel(fakeRequest) returns correct result")
}
