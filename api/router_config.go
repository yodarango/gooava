package api

import "keewee/repo"

var RouterConfig *repo.AppRepo

func SetRouterConfig(ar * repo.AppRepo){
	RouterConfig = ar
}