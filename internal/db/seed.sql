
SET FOREIGN_KEY_CHECKS = 0;

INSERT INTO users (first_name, last_name, email, password, created_at)
VALUES
('John', 'Doe', 'john.doe@example.com', 'password123', NOW()),
('Jane', 'Smith', 'jane.smith@example.com', 'password123', NOW()),
('Alice', 'Johnson', 'alice.johnson@example.com', 'password123', NOW());

INSERT INTO pantries (user_id, name, created_at)
VALUES
(1, 'Johns Pantry', NOW()),
(2, 'Janes Pantry', NOW()),
(3, 'Alices Pantry', NOW());

INSERT INTO ingredients (name)
VALUES
('Flour'),
('Sugar'),
('Eggs'),
('Milk'),
('Salt');

INSERT INTO pantry_ingredients (pantry_id, ingredient_id, quantity, created_at)
VALUES
(1, 1, '2 kg', NOW()),
(1, 2, '1 kg', NOW()),
(2, 3, '12 pieces', NOW()),
(2, 4, '1 liter', NOW()),
(3, 5, '500 g', NOW());

INSERT INTO prompts (user_id, description, provider, model, created_at)
VALUES
(1, 'Generate a healthy recipe using pantry ingredients.', 'OpenAI', 'GPT-4', NOW()),
(2, 'Create a quick dinner recipe.', 'OpenAI', 'GPT-4', NOW()),
(3, 'Find a budget-friendly dessert idea.', 'OpenAI', 'GPT-4', NOW());

INSERT INTO recipe_batches (user_id, name, recipe_count, is_healthy, is_quick, is_maximize_ingredients, is_budget_friendly, cuisine_type, prompt_id, created_at)
VALUES
(1, 'Healthy Meals', 3, TRUE, FALSE, TRUE, FALSE, 'Italian', 1, NOW()),
(2, 'Quick Dinners', 2, FALSE, TRUE, FALSE, TRUE, 'Mexican', 2, NOW()),
(3, 'Budget Desserts', 4, FALSE, FALSE, FALSE, TRUE, 'French', 3, NOW());

INSERT INTO recipes (user_id, batch_id, name, is_healthy, is_quick, is_maximize_ingredients, is_budget_friendly, cuisine_type, created_at, servings, instructions)
VALUES
(1, 1, 'Grilled Chicken Salad', TRUE, FALSE, TRUE, FALSE, 'Italian', NOW(), 4, 'Grill chicken, mix with salad ingredients.'),
(2, 2, 'Tacos', FALSE, TRUE, FALSE, TRUE, 'Mexican', NOW(), 4, 'Prepare taco shells and fill with meat and veggies.'),
(3, 3, 'Chocolate Cake', FALSE, FALSE, FALSE, TRUE, 'French', NOW(), 6, 'Mix ingredients, bake, and serve.');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, created_at)
VALUES
(1, 3, '500 g', NOW()),
(1, 4, '200 ml', NOW()),
(2, 5, '1 tsp', NOW()),
(2, 1, '300 g', NOW()),
(3, 2, '300 g', NOW()),
(3, 1, '500 g', NOW());

SET FOREIGN_KEY_CHECKS = 1;
