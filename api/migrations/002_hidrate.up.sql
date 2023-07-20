INSERT INTO api.package ("name",company,deed,drain,entry_type,entry,field_len) VALUES
	 ('three eyes',3,30,90,9,27,'{"deed": {"unit": 12, "title": 60}, "company": {"rn": 80, "tin": 80, "longname": 120}, "entry_type": {"code": 40, "unit": 12, "description": 120}}'),
	 ('one eye',1,10,30,3,9,'{"deed": {"unit": 3, "title": 15}, "company": {"rn": 20, "tin": 20, "longname": 30}, "entry_type": {"code": 10, "unit": 3, "description": 30}}'),
	 ('two eyes',2,20,60,6,18,'{"deed": {"unit": 6, "title": 30}, "company": {"rn": 40, "tin": 40, "longname": 60}, "entry_type": {"code": 20, "unit": 6, "description": 60}}');
INSERT INTO api."user" (id,"name",created_at,updated_at,email,api_key,package_kind,powers) VALUES
	 ('2PH24UhBlN5tlYdAmpdwiyPuWgB','Braila Crypto','2023-03-09 12:08:25+00','2023-03-09 12:08:25+00','braila.crypto@gmail.com','c88955c1c77689fe24201d2a05b43a583883409c54997926c54ced9cf747e18a','one eye','{create_own,write_own,read_own,delete_own}'),
	 ('2PH25DxmohuFCf3w73fQSTLJeVO','Gabriel Braila','2023-03-09 12:08:41+00','2023-03-09 12:08:41+00','braila.gabriel@gmail.com','ae07fb0572218441a7f8d123215d50919afe9878a4c006d5b152ddf99ecbac8b','two eyes','{create_own,write_own,read_own,delete_own}');
INSERT INTO api.auth (user_id,"source",source_id,access_token,refresh_token,expiry,created_at,updated_at) VALUES
	 ('2PH24UhBlN5tlYdAmpdwiyPuWgB','google','107467003969505231772','ya29.a0Ael9sCMhXSaW3G_n8T2Ooxj-5OXc2Yv7CJWkb4n8nkYPjZgz5neSLCIY8DmA8PLysjvPdHP-kMvfAsLAjQL3QK3M6Qix1HXMdM60NiWiPwHOigHTuWNF3DefmwGCgPjascPysTKKijoucBM0iG5eEsvTKjIOaCgYKAYISARESFQF4udJhf1qF1S9O-LWO2rr-4glaRw0163','','2023-04-02 17:53:49+00','2023-04-01 22:22:34+00','2023-04-02 16:53:50+00'),
	 ('2PH25DxmohuFCf3w73fQSTLJeVO','google','104793598026806295518','ya29.a0Ael9sCMhXSaW3G_n8T2Ooxj-5OXc2Yv7CJWkb4n8nkYPjZgz5neSLCIY8DmA8PLysjvPdHP-kMvfAsLAjQL3QK3M6Qix1HXMdM60NiWiPwHOigHTuWNF3DefmwGCgPjascPysTKKijoucBM0iG5eEsvTKjIOaCgYKAYISARESFQF4udJhf1qF1S9O-LWO2rr-4glaRw0163','','2023-04-02 17:53:49+00','2023-04-01 22:35:56+00','2023-04-02 16:53:50+00');
INSERT INTO api.company (tid,longname,tin,rn,deleted_at) VALUES
	 ('2PH24UhBlN5tlYdAmpdwiyPuWgB','China Work','CH37A46/1989','OPAwerww',NULL),
	 ('2PH24UhBlN5tlYdAmpdwiyPuWgB','China Workforce','CH39A46/1989','OPAwerww',NULL),
	 ('2PH25DxmohuFCf3w73fQSTLJeVO','volt-media','16728168','j40/14122/2004',NULL),
	 ('2PH24UhBlN5tlYdAmpdwiyPuWgB','tipografix','28728916','j12/24002/2008',NULL),
	 ('2PH25DxmohuFCf3w73fQSTLJeVO','ACME Enter','JP34546/2017','Acsd45gfr',NULL);
INSERT INTO api.deed (company_id,title,quantity,unit,unitprice,deleted_at) VALUES
	 (2,'pilule Nutrivit',150.0,'flacon100',0.87,NULL),
	 (2,'pita multa',12.0,'buc',2.15,NULL),
	 (2,'paine dospita',13.0,'buc',3.75,NULL),
	 (1,'pliante',150.0,'buc',0.01,'2023-03-12 08:53:18.672488+00'),
	 (1,'mape a5',100.0,'buc',0.05,'2023-04-20 08:11:23.227968+00'),
	 (1,'flyere A4/3',1111.0,'buc',0.01,'2023-04-20 08:12:41.183287+00'),
	 (2,'flyere a5',1500.0,'buc',0.01,NULL),
	 (2,'flyere A5/3',199.0,'buc',0.01,NULL),
	 (1,'flyere A4/3',1111.0,'buc',0.01,NULL),
	 (1,'Afise A3',777.0,'buc',0.01,NULL);
INSERT INTO api.deed (company_id,title,quantity,unit,unitprice,deleted_at) VALUES
	 (1,'afise',2000.0,'buc',0.01,NULL);
INSERT INTO api.entry_type (code,description,unit,tid,deleted_at) VALUES
	 ('hartie.a4.200g.mat','','buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('hartie.a4.200g.lucios','','buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('hartie.a2.110g.mat','','buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('hartie.a2.90g.mat','','buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('carton.sra3.300g.mat',NULL,'coala','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('stiplex.2x3.5mx5mm',NULL,'placa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('stiplex.1x2.2mx3mm',NULL,'placa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('plexic.2x5mx5mm',NULL,'placa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('plexic.1x2.2mx3mm',NULL,'placa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('rola.etichete.10x70mm.mat',NULL,'buc','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL);
INSERT INTO api.entry_type (code,description,unit,tid,deleted_at) VALUES
	 ('rola.etichete.50x50mm.lucios',NULL,'buc','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('autocolant.sra3.mat',NULL,'foi','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('autocolant.sra3.lucios',NULL,'foi','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('autocolant.pretaiat.sra3.lucios',NULL,'foaie','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('ps.sku-313','electric component to cope with temperature variation','piesa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('ps.sku-314',NULL,'piesa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('ps.sku-315',NULL,'piesa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('hartie.a4.110g.mat','cea mai eficienta hartie','buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('placa2x3.5m','atentie! se crapa usor','foaie','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('carton.sra2.200g.mat','se crapa la big','coala','2PH25DxmohuFCf3w73fQSTLJeVO',NULL);
INSERT INTO api.entry_type (code,description,unit,tid,deleted_at) VALUES
	 ('hartie.a4.90g.mat','nue suficient de opaca: tiparul pe ambele parti se vede','buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('rola.buline.lipire','bulinele sunt cam mici','rola','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('ps.sku-312','strange object','piesa','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('plexic.2x4.2.red',NULL,'placa','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('hartie.a4.90g.mat','hartie mata','pack.500','2PH24UhBlN5tlYdAmpdwiyPuWgB',NULL),
	 ('be.sat.119','benzi textile de satin de 119g','role','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('hartie.a4.110g.lucios','cea mai rezonabila hartie','buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('placa2x2.5mx2mm',NULL,'placa','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('be.sat.120','benzi textile de satin de 120g','rola','2PH25DxmohuFCf3w73fQSTLJeVO',NULL),
	 ('psr-404',NULL,'buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL);
INSERT INTO api.entry_type (code,description,unit,tid,deleted_at) VALUES
	 ('ps-405',NULL,'buc','2PH25DxmohuFCf3w73fQSTLJeVO',NULL);
INSERT INTO api.user_restriction (user_id,company,deed,drain,entry_type,entry,field_len) VALUES
	 ('2PH25DxmohuFCf3w73fQSTLJeVO',10,NULL,NULL,NULL,1,'{"deed": {"title": 100}}');
INSERT INTO api.entry (entry_type_id,date_added,quantity,company_id,deleted_at) VALUES
	 (10,'2022-02-07 08:48:07.74181+00',147.0,2,NULL),
	 (10,'2022-02-07 08:49:02.200706+00',487.0,2,NULL),
	 (2,'2023-03-15 10:48:52.476201+00',51.0,2,NULL),
	 (1,'2023-03-15 10:49:04.068565+00',751.0,2,NULL),
	 (3,'2023-03-15 10:49:20.813858+00',7.0,2,NULL),
	 (2,'2023-04-08 09:17:44.674374+00',101.0,2,NULL),
	 (14,'2023-03-15 10:49:29.582658+00',111.0,2,NULL),
	 (14,'2023-03-15 10:48:01.74239+00',111.0,2,NULL),
	 (2,'2022-02-06 10:28:02.796652+00',100.0,1,NULL),
	 (9,'2022-02-06 10:26:12.164527+00',200.0,1,NULL);
INSERT INTO api.entry (entry_type_id,date_added,quantity,company_id,deleted_at) VALUES
	 (7,'2022-02-06 10:26:12.164527+00',100.0,1,NULL),
	 (6,'2022-02-06 10:26:12.164527+00',100.0,1,NULL),
	 (8,'2022-02-06 10:45:09.800472+00',15.0,1,NULL),
	 (22,'2023-03-11 09:52:59.855666+00',7000.0,1,NULL),
	 (5,'2022-02-06 10:45:09.800472+00',200.0,1,NULL),
	 (7,'2022-02-06 10:28:02.796652+00',20.0,1,NULL),
	 (21,'2023-03-11 09:52:59.855666+00',1000.0,1,NULL),
	 (1,'2022-02-06 10:28:02.796652+00',100.0,1,NULL),
	 (2,'2023-03-15 10:45:48.335193+00',151.0,1,NULL),
	 (22,'2023-03-11 09:52:08.234907+00',5000.0,1,NULL);
INSERT INTO api.entry (entry_type_id,date_added,quantity,company_id,deleted_at) VALUES
	 (9,'2022-02-06 10:26:12.164527+00',200.0,1,NULL),
	 (3,'2022-02-06 10:28:02.796652+00',200.0,1,NULL),
	 (8,'2022-02-06 10:26:12.164527+00',200.0,1,NULL),
	 (21,'2023-03-11 09:52:08.234907+00',10000.0,1,NULL),
	 (4,'2023-04-20 07:36:40.419912+00',111.0,1,NULL),
	 (5,'2022-02-06 10:45:09.800472+00',100.0,1,NULL),
	 (3,'2022-02-06 10:45:09.800472+00',210.0,1,NULL),
	 (9,'2022-02-06 10:26:12.164527+00',200.0,1,NULL),
	 (6,'2022-02-06 10:45:09.800472+00',12.0,1,NULL),
	 (3,'2023-04-22 09:42:43+00',123.0,1,NULL);
INSERT INTO api.entry (entry_type_id,date_added,quantity,company_id,deleted_at) VALUES
	 (3,'2023-04-22 09:43:00+00',123.0,1,NULL),
	 (3,'2023-04-22 09:44:00+00',123.0,1,NULL),
	 (2,'2022-02-06 10:28:02.796652+00',12.0,1,NULL),
	 (4,'2022-02-06 10:45:09.800472+00',100.0,1,NULL),
	 (1,'2022-02-06 10:28:02.796652+00',200.0,1,NULL);
INSERT INTO api.drain (deed_id,entry_id,quantity,is_deleted) VALUES
	 (1,1,333.0,true),
	 (1,2,222.0,true),
	 (9,2,100.0,true),
	 (3,19,111.0,false),
	 (10,2,102.0,false),
	 (5,5,200.0,false),
	 (5,6,200.0,false),
	 (11,2,100.0,false);


