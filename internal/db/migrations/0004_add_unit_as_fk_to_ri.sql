ALTER TABLE recipe_ingredients
ADD CONSTRAINT fk_recipe_ingredients_measuring_unit 
    FOREIGN KEY (measuring_unit) 
    REFERENCES ingredient_units(name) 
    ON DELETE CASCADE;