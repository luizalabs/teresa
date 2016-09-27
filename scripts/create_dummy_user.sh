#!/bin/bash

# password: teresa123
echo 'insert into users (created_at, updated_at, name, email, password, is_admin) values ("2016-09-01 13:00:00-03:00", "2016-09-01 13:00:00-03:00", "admin", "admin@teresa.com", "$2a$10$RcJki6f/Bt.fbbtAS6ddh.05BzilrxRjWvd8ZJqpo9ToPbpiFY5.e", 1);' | sqlite3 teresa.sqlite
