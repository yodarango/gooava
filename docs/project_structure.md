### Project structure

```
/gooava/
├── /web/                      # All front-end code (HTML, CSS, JS, etc.)
│   ├── /templates/            # HTML templates (used by Go server)
│   │   ├── base.html          # Base layout (extends other templates)
│   │   ├── index.html         # Homepage template
│   │   ├── recipe.html        # Recipe details page
│   │   ├── /partials/         # Reusable HTML components
│   │   │   ├── header.html    # Header partial (e.g., nav bar)
│   │   │   ├── footer.html    # Footer partial
│   │   │   ├── recipe_card.html  # Recipe card component
│   │   │   └── modals/        # HTMX modal templates
│   │   │       └── edit_recipe_modal.html  # Modal for editing recipes
│   │   └── /forms/            # Forms for actions (add, edit, delete)
│   │       └── recipe_form.html  # Form for adding/editing a recipe
│   │
│   ├── /static/               # Static assets (CSS, JS, images)
│   │   ├── /css/              # CSS stylesheets
│   │   │   ├── app.css       # Global styles
│   │   │   ├── tokens.css    # Styles for recipe pages
│   │   └── /js/               # Optional JS (for HTMX extensions or minor enhancements)
│   │       ├── htmx_extensions.js  # JS for handling HTMX-specific behaviors
│   │   └── /images/           # Static images (e.g., recipe images, icons)
│   │       ├── logo.png       # App logo
│   │       └── example.jpg    # Example recipe image
│   │
│   └── favicon.ico            # Favicon for the website
│
├── /api/                   # API layer for handling HTTP requests and responses
│   ├── /v1/                # Versioned API endpoints
│   │   ├── recipe.go       # Handlers for recipe-related operations
│   │   ├── user.go         # Handlers for user-related operations
│   │   ├── cart.go         # Handlers for cart-related operations
│   │   └── middleware.go   # Middleware (e.g., authentication, logging)
│   ├── routes.go           # API route definitions
│   └── response.go         # Utilities for building consistent API responses
│
├── /cmd/                   # Main applications for different services
│   └── recipe-app/         # The main service application
│       └── main.go         # Application entry point
│
├── /db/                    # Database-related files
│   ├── /migrations/        # SQL files for database schema and migrations
│   │   ├── 001_create_users_table.sql
│   │   ├── 002_create_recipes_table.sql
│   │   ├── 003_create_prompts_table.sql
│   │   ├── 004_create_ingredients_table
|   |   ├── 005_create_recipe_ingredients_table.sql
│   └── migrate.sh
|
├── /config/                # Configuration files for different environments
│   ├── config.go           # Configuration handling (e.g., environment variables)
│   ├── config_test.go      # Tests for config
│   └── app.json            # App configuration (e.g., DB settings, API keys)
│
├── /docs/                  # Documentation for the project (e.g., API specs, design decisions)
│
├── /internal/              # Non-public application and library code
│   ├── /models/            # Application-specific domain logic
│   │   ├── recipe.go       # Recipe domain model (business logic)
│   │   ├── user.go         # User domain model (business logic)
│   ├── /db/                # Database-related code (repositories)
│   │   ├── recipe_repo.go  # Repository for recipe data
│   │   ├── user_repo.go    # Repository for user data
│   │   └── db.go           # Database connection logic
│   ├── /ai/                # Logic for calling and handling AI prompts
│   │   └── prompt.go       # Prompt creation and handling logic
│   ├── /grocery/           # Logic for interacting with external grocery services
│   │   ├── puppeteer.go    # Puppeteer service for adding items to cart
│   │   └── api.go          # API service for adding items to cart
│   ├── /utils/             # Utility functions (helpers, string processing, etc.)
│   │   └── utils.go        # General utility functions
│   └── /auth/              # Authentication logic (if required)
│       └── auth.go         # Authentication middleware and helper functions
│
├── /pkg/                   # Shared code that can be used by other projects
│   ├── /logger/            # Logging utilities
│   │   └── logger.go       # Centralized logging logic
│   └── /http/              # HTTP helpers (e.g., custom error handling)
│       └── http_helper.go  # HTTP-related helper functions
│
├── /scripts/               # Scripts for automation tasks (e.g., database migrations)
│   └── migrate.sh          # Example migration script
│
├── /test/                  # Testing utilities
│   ├── /mocks/             # Mock data for testing
│   ├── /integration/       # Integration tests
│   └── /unit/              # Unit tests for individual components
│       ├── recipe_test.go  # Tests for recipe domain logic
│       └── user_test.go    # Tests for user domain logic
│
├── /vendor/                # Vendor directory for dependencies (if vendoring is used)
│
├── go.mod                  # Go modules file
├── go.sum                  # Go dependencies file
├── Dockerfile              # Dockerfile for containerizing the application
└── Makefile                # Makefile for automating tasks
```
