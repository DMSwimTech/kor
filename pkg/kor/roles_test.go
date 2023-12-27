package kor

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func createTestRoles(clientset *fake.Clientset, t *testing.T) *fake.Clientset {

	role1 := CreateTestRole(testNamespace, "test-role1")
	role2 := CreateTestRole(testNamespace, "test-role2")
	_, err := clientset.RbacV1().Roles(testNamespace).Create(context.TODO(), role1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "Role", err)
	}

	_, err = clientset.RbacV1().Roles(testNamespace).Create(context.TODO(), role2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "Role", err)
	}

	testRoleRef := CreateTestRoleRef("test-role1")
	testRoleBinding := CreateTestRoleBinding(testNamespace, "test-rb", "test-sa", testRoleRef)
	_, err = clientset.RbacV1().RoleBindings(testNamespace).Create(context.TODO(), testRoleBinding, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "Role", err)
	}

	return clientset
}

func createTestRolesClient(t *testing.T) *fake.Clientset {
	clientset := fake.NewSimpleClientset()

	_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
	}, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("Error creating namespace %s: %v", testNamespace, err)
	}

	createTestRoles(clientset, t)

	return clientset
}

func TestRetrieveUsedRoles(t *testing.T) {
	clientset := createTestRolesClient(t)

	usedRoles, err := retrieveUsedRoles(clientset, testNamespace, &FilterOptions{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(usedRoles) != 1 {
		t.Errorf("Expected 1 used role, got %d", len(usedRoles))
	}

	if usedRoles[0] != "test-role1" {
		t.Errorf("Expected 'test-role1', got %s", usedRoles[0])
	}
}

func TestRetrieveRoleNames(t *testing.T) {
	clientset := createTestRolesClient(t)
	allRoles, err := retrieveRoleNames(clientset, testNamespace)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(allRoles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(allRoles))
	}
}

func TestProcessNamespaceRoles(t *testing.T) {
	clientset := createTestRolesClient(t)

	unusedRoles, err := processNamespaceRoles(clientset, testNamespace, &FilterOptions{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(unusedRoles) != 1 {
		t.Errorf("Expected 1 unused role, got %d", len(unusedRoles))
	}

	if unusedRoles[0] != "test-role2" {
		t.Errorf("Expected 'test-role2', got %s", unusedRoles[0])
	}
}

func TestGetUnusedRolesStructured(t *testing.T) {
	clientset := createTestRolesClient(t)

	includeExcludeLists := IncludeExcludeLists{
		IncludeListStr: "",
		ExcludeListStr: "",
	}

	opts := Opts{
		WebhookURL:    "",
		Channel:       "",
		Token:         "",
		DeleteFlag:    false,
		NoInteractive: true,
	}

	output, err := GetUnusedRoles(includeExcludeLists, &FilterOptions{}, clientset, "json", opts)
	if err != nil {
		t.Fatalf("Error calling GetUnusedRolesStructured: %v", err)
	}

	expectedOutput := map[string]map[string][]string{
		testNamespace: {
			"Roles": {"test-role2"},
		},
	}

	var actualOutput map[string]map[string][]string
	if err := json.Unmarshal([]byte(output), &actualOutput); err != nil {
		t.Fatalf("Error unmarshaling actual output: %v", err)
	}

	if !reflect.DeepEqual(expectedOutput, actualOutput) {
		t.Errorf("Expected output does not match actual output")
	}
}

func init() {
	scheme.Scheme = runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme.Scheme)
}
