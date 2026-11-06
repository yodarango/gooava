-- Disabilita i controlli delle chiavi esterne
SET FOREIGN_KEY_CHECKS = 0;

-- Aggiunta delle colonne batch_id e measuring_unit, modifica di quantity e aggiunta delle chiavi esterne
ALTER TABLE pantry_ingredients
ADD COLUMN measuring_unit VARCHAR(50) NOT NULL, -- Aggiunta della colonna measuring_unit
MODIFY COLUMN quantity INT(5) NOT NULL, -- Modifica della colonna quantity
ADD CONSTRAINT fk_pantry_ingredients_measuring_unit 
    FOREIGN KEY (measuring_unit) 
    REFERENCES ingredient_units(name) 
    ON DELETE CASCADE;

-- Riabilita i controlli delle chiavi esterne
SET FOREIGN_KEY_CHECKS = 1;
