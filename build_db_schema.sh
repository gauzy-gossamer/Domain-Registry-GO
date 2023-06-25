#!/bin/bash
# printing-out sql command in right order to create of new database
#base system

set -e

DIR=$(dirname $0)/sql

write_script() 
{
    cat $DIR/array_util_func.sql
    cat $DIR/error.sql
    cat $DIR/enum_reason.sql
    cat $DIR/enum_ssntype.sql
    cat $DIR/enum_country.sql
    cat $DIR/zone.sql
    #registar and registraracl  tables
    cat $DIR/registrar.sql
    # object table
    cat $DIR/history_base.sql
    cat $DIR/ccreg.sql
    cat $DIR/history.sql
    #zone generator
    cat $DIR/genzone.sql
    #adif
    cat $DIR/admin.sql  
    # banking
    cat $DIR/enum_bank_code.sql
    cat $DIR/credit_ddl.sql
    cat $DIR/invoice.sql
    cat $DIR/bank.sql
    cat $DIR/bank_ddl_new.sql
    # common functions
    cat $DIR/func.sql
    # table with parameters
    cat $DIR/enum_params.sql
    # keyset
    cat $DIR/keyset.sql
    # state and poll
    cat $DIR/state.sql
    cat $DIR/poll.sql
    # notify mailer
    cat $DIR/notify_new.sql
    # new table with filters
    cat $DIR/filters.sql
    # new indexes for history tables
    cat $DIR/index.sql
    # registrar's certifications and groups
    cat $DIR/registrar_certification_ddl.sql
    cat $DIR/registrar_disconnect.sql
    # monitoring
    cat $DIR/monitoring_dml.sql
    # epp login IDs
    cat $DIR/epp_login.sql
    # changes notifications
    cat $DIR/changes_notifications.sql
    # domain name validators
    cat $DIR/enum_domain_name_validation_checker_dml.sql
    # Registrar Notifications
    cat $DIR/registrar_notify.sql
    # DNSSEC Services
    cat $DIR/dnssec.sql
    # Zone Responsible Registrar
    cat $DIR/zone_responsible_registrar.sql
}

logger()
{
    cat $DIR/logger_ddl.sql		
    cat $DIR/logger_dml_whois.sql
    cat $DIR/logger_dml_epp.sql
    cat $DIR/logger_dml.sql
    cat $DIR/logger_dml_admin.sql
    cat $DIR/logger_dml_rdap.sql
    cat $DIR/logger_dml_regadmin.sql
    cat $DIR/logger_partitioning.sql
}

usage()
{
    echo "$0 : Create database installation .sql script. It accepts one of these options: "
    echo "		--without-log exclude logging tables (used by fred-logd daemon) "
    echo "		--help 	   display this message "
}

case "$1" in
    --without-log)
            write_script
            ;;
    --with-test)
            write_script
            logger
            cat $DIR/test.sql
            ;;
    --help) 
            usage
            ;;
    *)
            write_script
            logger
            ;;
esac
