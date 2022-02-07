package loginrecord

import (
	"k8s.io/apimachinery/pkg/runtime"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"

	"aiscope/pkg/api"
	"aiscope/pkg/apiserver/query"
	aiInformers "aiscope/pkg/client/informers/externalversions"
	"aiscope/pkg/models/resources/v1alpha2"
)

const recordType = "type"

type loginrecordsGetter struct {
	aiInformer aiInformers.SharedInformerFactory
}

func New(aiInformer aiInformers.SharedInformerFactory) v1alpha2.Interface {
	return &loginrecordsGetter{aiInformer: aiInformer}
}

func (d *loginrecordsGetter) Get(_, name string) (runtime.Object, error) {
	return d.aiInformer.Iam().V1alpha2().Users().Lister().Get(name)
}

func (d *loginrecordsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	records, err := d.aiInformer.Iam().V1alpha2().LoginRecords().Lister().List(query.Selector())

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, user := range records {
		result = append(result, user)
	}

	return v1alpha2.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *loginrecordsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftUser, ok := left.(*iamv1alpha2.LoginRecord)
	if !ok {
		return false
	}

	rightUser, ok := right.(*iamv1alpha2.LoginRecord)
	if !ok {
		return false
	}

	return v1alpha2.DefaultObjectMetaCompare(leftUser.ObjectMeta, rightUser.ObjectMeta, field)
}

func (d *loginrecordsGetter) filter(object runtime.Object, filter query.Filter) bool {
	record, ok := object.(*iamv1alpha2.LoginRecord)

	if !ok {
		return false
	}

	switch filter.Field {
	case recordType:
		return string(record.Spec.Type) == string(filter.Value)
	default:
		return v1alpha2.DefaultObjectMetaFilter(record.ObjectMeta, filter)
	}

}

