UPDATE donations SET donator=95538921, receiver=2114454465, "timestamp"='2018-08-25 12:17:37', note='hello, world', amount=1337 WHERE transaction_id=1000;
INSERT INTO donations (transaction_id, donator, receiver, "timestamp", note, amount)
       SELECT 1000, 95538921, 2114454465, '2018-08-25 12:17:37', 'hello, world', 1337
       WHERE NOT EXISTS (SELECT 1 FROM donations WHERE transaction_id=1000);
