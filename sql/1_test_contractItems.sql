UPDATE contractItems SET contract_id=1000, type_id=42, item_id=10000, quantity=100 WHERE id=1;
INSERT INTO contractItems (id, contract_id, type_id, item_id, quantity)
       SELECT 1, 1000, 42, 10000, 100
       WHERE NOT EXISTS (SELECT 1 FROM contractItems WHERE id=1);

UPDATE contractItems SET contract_id=1000, type_id=587, item_id=10000, quantity=1 WHERE id=2;
INSERT INTO contractItems (id, contract_id, type_id, item_id, quantity)
       SELECT 2, 1000, 587, 10000, 1
       WHERE NOT EXISTS (SELECT 1 FROM contractItems WHERE id=2);
