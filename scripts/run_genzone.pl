#! /usr/bin/perl

use strict;

my $rndc = 'rndc';
my $python3 = 'python3';

`$python3 generate_zone.py --run-named-checkzone --target-dir /etc/bind/`;
die $! if $?;

foreach my $zonefile (</etc/bind/*db>) {
    next if $zonefile !~ /([\w\-\.]+)\.db$/;
    my $zone = $1;
    my $ret = `$rndc zonestatus $zone 2>&1`;
    if ($ret =~ /no matching zone/) {
        `$rndc addzone $zone '{type master; file "$zonefile";};'`;
        die $! if $?;
    }

    `$rndc reload $zone`;
    die $! if $?;
    print "loaded $zone\n";
}
