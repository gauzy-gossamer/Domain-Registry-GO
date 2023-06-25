INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (2, 'dncheck_no_consecutive_hyphens', 'forbid consecutive hyphens');
INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (5, 'dncheck_not_empty_domain_name', 'forbid empty domain name');
INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (6, 'dncheck_rfc1035_preferred_syntax', 'enforces rfc1035 preferred syntax');
INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (7, 'dncheck_single_digit_labels_only', 'enforces single digit labels (for enum domains)');
INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (8, 'dncheck_no_idn_punycode', 'forbid idn punycode encoding');
INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (9, 'dncheck_no_single_character', 'forbid single character domains');
INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (10, 'dncheck_no_34_hyphens', 'forbid hyphens in domains at 3d and 4th position');
INSERT INTO enum_domain_name_validation_checker (id, name, description)
    VALUES (11, 'dncheck_su_idn', 'validity of idn domains for su zone');

