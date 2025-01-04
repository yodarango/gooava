package models

type ResponseError struct {
	Code  string      `json:"code"`
	Error interface{} `json:"error"`
	Title string      `json:"title"`
}
