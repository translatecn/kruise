package validating

import (
	"fmt"
	"strconv"
	"testing"

	"k8s.io/apimachinery/pkg/util/uuid"

	appsv1alpha1 "github.com/openkruise/kruise/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type testCase struct {
	spec    *appsv1alpha1.CloneSetSpec
	oldSpec *appsv1alpha1.CloneSetSpec
}

func TestValidate(t *testing.T) {
	validLabels := map[string]string{"a": "b"}
	validPodTemplate := v1.PodTemplate{
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validLabels,
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
				Containers:    []v1.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
			},
		},
	}
	validPodTemplate1 := v1.PodTemplate{
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validLabels,
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
				Containers:    []v1.Container{{Name: "abc", Image: "image1", ImagePullPolicy: "IfNotPresent"}},
			},
		},
	}
	validPodTemplate2 := v1.PodTemplate{
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validLabels,
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
				Containers:    []v1.Container{{Name: "abc", Image: "image2", ImagePullPolicy: "Always"}},
			},
		},
	}

	invalidLabels := map[string]string{"NoUppercaseOrSpecialCharsLike=Equals": "b"}
	invalidPodTemplate := v1.PodTemplate{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: invalidLabels,
			},
		},
	}

	var valTrue = true
	var val1 int32 = 1
	var val2 int32 = 2
	var minus1 int32 = -1
	maxUnavailable0 := intstr.FromInt(0)
	maxUnavailable1 := intstr.FromInt(1)
	maxUnavailable120Percent := intstr.FromString("120%")

	uid := uuid.NewUUID()
	p0 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "p0",
			Namespace:       metav1.NamespaceDefault,
			OwnerReferences: []metav1.OwnerReference{{UID: uid, Controller: &valTrue}},
		},
	}

	successCases := []testCase{
		{
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
				ScaleStrategy: appsv1alpha1.CloneSetScaleStrategy{
					PodsToDelete: []string{"p0"},
				},
			},
		},
		{
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable120Percent,
				},
			},
		},
		{
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable120Percent,
					PriorityStrategy: &appsv1alpha1.UpdatePriorityStrategy{
						WeightPriority: []appsv1alpha1.UpdatePriorityWeightTerm{
							{Weight: 20, MatchSelector: metav1.LabelSelector{MatchLabels: map[string]string{"key": "foo"}}},
						},
					},
				},
			},
		},
		{
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
			oldSpec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate1.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			obj := appsv1alpha1.CloneSet{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("cs-%d", i), Namespace: metav1.NamespaceDefault, UID: uid, ResourceVersion: "2"},
				Spec:       *successCase.spec,
			}
			h := CloneSetCreateUpdateHandler{Client: fake.NewFakeClient(&p0)}
			if successCase.oldSpec == nil {
				if errs := h.validateCloneSet(&obj); len(errs) != 0 {
					t.Errorf("expected success: %v", errs)
				}
			} else {
				oldObj := appsv1alpha1.CloneSet{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("cs-%d", i), Namespace: metav1.NamespaceDefault, UID: uid, ResourceVersion: "1"},
					Spec:       *successCase.oldSpec,
				}
				if errs := h.validateCloneSetUpdate(&obj, &oldObj); len(errs) != 0 {
					t.Errorf("expected success: %v", errs)
				}
			}
		})
	}

	errorCases := map[string]testCase{
		"invalid-replicas": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &minus1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
		"invalid-template": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: invalidPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
		"invalid-selector": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "b", "c": "d"}},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
		"invalid-update-type": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           "",
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
		"invalid-partition": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &minus1,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
		"invalid-maxUnavailable": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable0,
				},
			},
		},
		"invalid-podsToDelete-1": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
				ScaleStrategy: appsv1alpha1.CloneSetScaleStrategy{
					PodsToDelete: []string{"p0", "p0"},
				},
			},
		},
		"invalid-podsToDelete-2": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
				ScaleStrategy: appsv1alpha1.CloneSetScaleStrategy{
					PodsToDelete: []string{"p0", "p1"},
				},
			},
		},
		"invalid-cloneset-update-1": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas:             &val1,
				Selector:             &metav1.LabelSelector{MatchLabels: validLabels},
				RevisionHistoryLimit: &val2,
				Template:             validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
			oldSpec: &appsv1alpha1.CloneSetSpec{
				Replicas:             &val1,
				Selector:             &metav1.LabelSelector{MatchLabels: validLabels},
				RevisionHistoryLimit: &val1,
				Template:             validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
		"invalid-cloneset-update-2": {
			spec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
			oldSpec: &appsv1alpha1.CloneSetSpec{
				Replicas: &val1,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				Template: validPodTemplate2.Template,
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
					Type:           appsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
					Partition:      &val2,
					MaxUnavailable: &maxUnavailable1,
				},
			},
		},
	}

	for k, v := range errorCases {
		t.Run(k, func(t *testing.T) {
			obj := appsv1alpha1.CloneSet{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("cs-%v", k), Namespace: metav1.NamespaceDefault, UID: uid, ResourceVersion: "2"},
				Spec:       *v.spec,
			}
			h := CloneSetCreateUpdateHandler{Client: fake.NewFakeClient(&p0)}
			if v.oldSpec == nil {
				if errs := h.validateCloneSet(&obj); len(errs) == 0 {
					t.Errorf("expected failure for %v", k)
				}
			} else {
				oldObj := appsv1alpha1.CloneSet{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("cs-%v", k), Namespace: metav1.NamespaceDefault, UID: uid, ResourceVersion: "1"},
					Spec:       *v.oldSpec,
				}
				if errs := h.validateCloneSetUpdate(&obj, &oldObj); len(errs) == 0 {
					t.Errorf("expected failure for %v", k)
				}
			}
		})
	}
}