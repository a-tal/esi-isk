UPDATE contracts SET donator=95538921, receiver=2114454465, location=60014842, issued='2018-08-25 06:17:37', expires='2018-08-27 06:17:37', accepted='0', system=30000380 WHERE contract_id=1000;
INSERT INTO contracts (contract_id, donator, receiver, location, issued, expires, accepted, system)
       SELECT 1000, 95538921, 2114454465, 60014842, '2018-08-25 06:17:37', '2018-08-27 06:17:37', '0', 30000380
       WHERE NOT EXISTS (SELECT 1 FROM contracts WHERE contract_id=1000);
