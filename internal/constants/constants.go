package constants

const ENV_PROD = "PROD"
const ENV_DEV = "DEV"

// routes
const ROUTE_ROOT = "/"

// /recipe-batch/*
const ROUTE_RECIPEBATCHES = ROUTE_ROOT + "recipe-batches"
const ROUTE_RECIPE_BATCHES_NEW = ROUTE_RECIPEBATCHES + "/new"
const ROUTE_RECIPEBATCHES_ID = ROUTE_RECIPEBATCHES + "/:id"
const ROUTE_RECIPEBATCHES_ID_INGREDIENTS = ROUTE_RECIPEBATCHES_ID + "/ingredients"

// /recipes/*
const ROUTE_RECIPES = ROUTE_ROOT + "recipes"
