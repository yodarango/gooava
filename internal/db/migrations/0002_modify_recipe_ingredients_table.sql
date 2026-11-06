-- Disabilita i controlli delle chiavi esterne
SET FOREIGN_KEY_CHECKS = 0;

-- Aggiunta della colonna batch_id e measuring_unit, e modifica di quantity a INT(5)
ALTER TABLE recipe_ingredients
ADD COLUMN batch_id INT UNSIGNED NOT NULL,
ADD COLUMN measuring_unit VARCHAR(50) NOT NULL,
MODIFY COLUMN quantity INT(5) NOT NULL,
ADD CONSTRAINT fk_recipe_ingredients_batch_id FOREIGN KEY (batch_id) REFERENCES recipe_batches(id) ON DELETE CASCADE;

-- Riabilita i controlli delle chiavi esterne
SET FOREIGN_KEY_CHECKS = 1;
