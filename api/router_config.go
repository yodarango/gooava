package api

import "gooava/repo"

var RouterConfig *repo.AppRepo

func SetRouterConfig(ar * repo.AppRepo){
	RouterConfig = ar
}