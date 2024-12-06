/*
Copyright 2024-2026 ECCO Data & AI Open-Source Project Maintainers.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	sciencev1 "science.sneaksanddata.com/nexus-core/pkg/apis/science/v1"
	versioned "science.sneaksanddata.com/nexus-core/pkg/generated/clientset/versioned"
	internalinterfaces "science.sneaksanddata.com/nexus-core/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "science.sneaksanddata.com/nexus-core/pkg/generated/listers/science/v1"
)

// MachineLearningAlgorithmInformer provides access to a shared informer and lister for
// MachineLearningAlgorithms.
type MachineLearningAlgorithmInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.MachineLearningAlgorithmLister
}

type machineLearningAlgorithmInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMachineLearningAlgorithmInformer constructs a new informer for MachineLearningAlgorithm type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMachineLearningAlgorithmInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMachineLearningAlgorithmInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMachineLearningAlgorithmInformer constructs a new informer for MachineLearningAlgorithm type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMachineLearningAlgorithmInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ScienceV1().MachineLearningAlgorithms(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ScienceV1().MachineLearningAlgorithms(namespace).Watch(context.TODO(), options)
			},
		},
		&sciencev1.MachineLearningAlgorithm{},
		resyncPeriod,
		indexers,
	)
}

func (f *machineLearningAlgorithmInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMachineLearningAlgorithmInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *machineLearningAlgorithmInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&sciencev1.MachineLearningAlgorithm{}, f.defaultInformer)
}

func (f *machineLearningAlgorithmInformer) Lister() v1.MachineLearningAlgorithmLister {
	return v1.NewMachineLearningAlgorithmLister(f.Informer().GetIndexer())
}
