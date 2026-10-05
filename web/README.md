# Gooava

A full home management app built with Go and React — from finances to recipes, everything your household needs in one place.

## Features

- **Finances** — track expenses and incomes, manage recurring bills, and plan spending around your paychecks
- **Bank sync** — connect your bank via Plaid to sync balances and transactions automatically
- **Calendar** — see paydays, bill due dates, and household events at a glance
- **Meals & recipes** — plan meals and keep your favorite recipes (coming soon)
- **Accounts** — email verification, welcome, and password reset flows

## Stack

- Go backend with JWT authentication and MySQL (migrations included)
- React frontend with Vite, React Router, and a custom design system
- Docker and Docker Compose configuration

## NAMING CONVENTIONS

- All dates prefixed with at are in `time.Time` format
- All dates suffixed with `Date` are in `string` format
- Function that return Date structs must return the date `2000-01-01 00:00:00` if there is no date to return. This is to avoid returning a `&reference` of date to allow for `nil`
