UPDATE characters SET character_name='fake name 1', corporation_id=109299958, corporation_name='fake corp 7', received=1, received_isk=123213213.5, donated=0, donated_isk=0, joined='2018-08-25 06:17:37',  last_seen='2018-08-25 06:27:37', last_received='2018-08-26 08:27:37' WHERE character_id=95538921;
INSERT INTO characters (character_id, character_name, corporation_id, corporation_name, received, received_isk, donated, donated_isk, joined, last_seen, last_received)
       SELECT 95538921, 'fake name 1', 109299958, 'fake corp 1', 1, 123213213.5, 0, 0, '2018-08-25 06:17:37', '2018-08-25 06:27:37', '2018-08-26 08:27:37'
       WHERE NOT EXISTS (SELECT 1 FROM characters WHERE character_id=95538921);


UPDATE characters SET character_name='fake name 2', corporation_id=924269309, corporation_name='fake corp 7', received=4, received_isk=12313213.5, donated=0, donated_isk=0, joined='2018-08-25 07:17:37',  last_seen='2018-08-25 08:27:37', last_received='2018-08-26 03:27:37' WHERE character_id=2113116864;
INSERT INTO characters (character_id, character_name, corporation_id, corporation_name, received, received_isk, donated, donated_isk, joined, last_seen, last_received)
       SELECT 2113116864, 'fake name 2', 924269309, 'fake corp 2', 4, 12313213.5, 0, 0, '2018-08-25 07:17:37', '2018-08-25 08:27:37', '2018-08-26 03:27:37'
       WHERE NOT EXISTS (SELECT 1 FROM characters WHERE character_id=2113116864);


UPDATE characters SET character_name='fake name 3', corporation_id=924269309, corporation_name='fake corp 7', received=9, received_isk=123213.5, donated=0, donated_isk=0, joined='2018-08-26 07:17:37',  last_seen='2018-08-26 08:27:37', last_received='2018-08-27 03:27:37' WHERE character_id=238692228;
INSERT INTO characters (character_id, character_name, corporation_id, corporation_name, received, received_isk, donated, donated_isk, joined, last_seen, last_received)
       SELECT 238692228, 'fake name 3', 924269309, 'fake corp 3', 9, 123213.5, 0, 0, '2018-08-26 07:17:37', '2018-08-26 08:27:37', '2018-08-27 03:27:37'
       WHERE NOT EXISTS (SELECT 1 FROM characters WHERE character_id=238692228);


UPDATE characters SET character_name='fake name 4', corporation_id=98064493, corporation_name='fake corp 7', received=0, received_isk=0, donated=5, donated_isk=1232131.1, joined='2018-08-24 07:17:37',  last_seen='2018-08-24 08:27:37', last_donated='2018-08-25 01:27:37' WHERE character_id=1418436663;
INSERT INTO characters (character_id, character_name, corporation_id, corporation_name, received, received_isk, donated, donated_isk, joined, last_seen, last_donated)
       SELECT 1418436663, 'fake name 4', 98064493, 'fake corp 4', 0, 0, 5, 1232131.1, '2018-08-24 07:17:37', '2018-08-24 08:27:37', '2018-08-25 01:27:37'
       WHERE NOT EXISTS (SELECT 1 FROM characters WHERE character_id=1418436663);


UPDATE characters SET character_name='fake name 5', corporation_id=98571038, corporation_name='fake corp 7', received=0, received_isk=0, donated=15, donated_isk=1232213131.1, joined='2018-08-24 01:17:37',  last_seen='2018-08-24 01:27:37', last_donated='2018-08-25 09:27:37' WHERE character_id=95535526;
INSERT INTO characters (character_id, character_name, corporation_id, corporation_name, received, received_isk, donated, donated_isk, joined, last_seen, last_donated)
       SELECT 95535526, 'fake name 5', 98571038, 'fake corp 5', 0, 0, 15, 1232213131.1, '2018-08-24 01:17:37', '2018-08-24 01:27:37', '2018-08-25 09:27:37'
       WHERE NOT EXISTS (SELECT 1 FROM characters WHERE character_id=95535526);


UPDATE characters SET character_name='fake name 6', corporation_id=98224639, corporation_name='fake corp 7', received=0, received_isk=0, donated=42, donated_isk=91232123131.15, joined='2018-08-21 01:17:37',  last_seen='2018-08-21 01:27:37', last_donated='2018-08-22 09:27:37' WHERE character_id=605167438;
INSERT INTO characters (character_id, character_name, corporation_id, corporation_name, received, received_isk, donated, donated_isk, joined, last_seen, last_donated)
       SELECT 605167438, 'fake name 6', 98224639, 'fake corp 6', 0, 0, 42, 91232123131.15, '2018-08-21 01:17:37', '2018-08-21 01:27:37', '2018-08-22 09:27:37'
       WHERE NOT EXISTS (SELECT 1 FROM characters WHERE character_id=605167438);


UPDATE characters SET character_name='fake name 7', corporation_id=1000077, corporation_name='fake corp 7', received=0, received_isk=0, donated=12, donated_isk=4123223131.32, joined='2018-08-21 02:17:37',  last_seen='2018-08-21 11:27:37', last_donated='2018-08-22 14:27:37' WHERE character_id=2114454465;
INSERT INTO characters (character_id, character_name, corporation_id, corporation_name, received, received_isk, donated, donated_isk, joined, last_seen, last_donated)
       SELECT 2114454465, 'fake name 7', 1000077, 'fake corp 7', 0, 0, 12, 4123223131.32, '2018-08-21 02:17:37', '2018-08-21 11:27:37', '2018-08-22 14:27:37'
       WHERE NOT EXISTS (SELECT 1 FROM characters WHERE character_id=2114454465);
