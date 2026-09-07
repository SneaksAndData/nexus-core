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
		AppliedConfiguration:    "b64__eyJjb250YWluZXIiOnsiaW1hZ2UiOiJ0ZXN0LmlvIiwicmVnaXN0cnkiOiJhbGdvcml0aG1zL3Rlc3QiLCJ2ZXJzaW9uVGFnIjoidjEuMC4wIiwic2VydmljZUFjY291bnROYW1lIjoidGVzdC1zYSJ9LCJjb21wdXRlUmVzb3VyY2VzIjp7ImNwdUxpbWl0IjoiMTAwMG0iLCJtZW1vcnlMaW1pdCI6IjIwMDBNaSJ9LCJ3b3JrZ3JvdXBSZWYiOnsibmFtZSI6InRlc3Qtd29ya2dyb3VwIiwiZ3JvdXAiOiJuZXh1cy13b3JrZ3JvdXAuaW8iLCJraW5kIjoiS2FycGVudGVyV29ya2dyb3VwVjEifSwiY29tbWFuZCI6InB5dGhvbiIsImFyZ3MiOlsiam9iLnB5IiwiLS1zYXMtdXJpPSVzIiwiLS1yZXF1ZXN0LWlkPSVzIiwiLS1hcmcxPXRydWUiXSwicGF5bG9hZENvbmZpZ3VyYXRpb24iOnsicGF5bG9hZFZhbGlkRm9yIjoiMjRoIiwicGF5bG9hZFNlcmlhbGl6YXRpb25Nb2RlIjoiYmFja2VuZCJ9LCJydW50aW1lRW52aXJvbm1lbnQiOnsiZGVhZGxpbmVTZWNvbmRzIjoxMjAsIm1heGltdW1SZXRyaWVzIjozfSwiZGF0YWRvZ0ludGVncmF0aW9uU2V0dGluZ3MiOnsibW91bnREYXRhZG9nU29ja2V0Ijp0cnVlfX0=",
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
