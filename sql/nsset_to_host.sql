CREATE TABLE public.domain_host_map (
    domainid integer NOT NULL, 
    hostid integer NOT NULL
);

CREATE INDEX domain_host_map_idx ON public.domain_host_map USING btree (domainid);
CREATE INDEX domain_host_map_idx2 ON public.domain_host_map USING btree (hostid);

ALTER TABLE ONLY public.domain_host_map
    ADD CONSTRAINT domain_hostid_fkey FOREIGN KEY (hostid) REFERENCES public.host(hostid) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY public.domain_host_map
    ADD CONSTRAINT domain_host_domainid_fkey FOREIGN KEY (domainid) REFERENCES public.domain(id) ON UPDATE CASCADE ON DELETE CASCADE;



CREATE TABLE public.host (
    hostid integer NOT NULL,
    fqdn character varying(255) NOT NULL,
    delegatable boolean default true
);

ALTER TABLE ONLY public.host_ipaddr_map
    ADD CONSTRAINT host_hostid_fkey FOREIGN KEY (hostid) REFERENCES public.host(hostid) ON UPDATE CASCADE ON DELETE CASCADE;

CREATE UNIQUE INDEX host_idx ON public.host USING btree (hostid);

ALTER TABLE ONLY public.host
    ADD CONSTRAINT host_hostid_fkey FOREIGN KEY (hostid) REFERENCES public.object_registry(id) ON UPDATE CASCADE ON DELETE CASCADE;


CREATE VIEW public.domain_states AS
 SELECT d.id AS object_id,
    o.historyid AS object_hid,
    (((((COALESCE(osr.states, '{}'::integer[]) ||
        CASE
            WHEN public.timestamp_test(d.exdate, '0'::character varying, '0'::character varying, 'UTC'::character varying) THEN ARRAY[9, 1, 3]
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN (d.nsset IS NULL) THEN ARRAY[14]
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN (NOT public.date_time_test((d.exdate)::date, ep_ex_not.val, '0'::character varying, ep_tz.val)) THEN ARRAY[2]
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN (((   SELECT count(*) AS hostsn
               FROM (public.host h
                 JOIN public.domain_host_map dhm ON (((h.hostid = dhm.hostid) AND
                    CASE
                        WHEN public.is_subordinate(lower((h.fqdn)::text), lower((o.name)::text)) THEN (SELECT count(*) 
                            FROM host_ipaddr_map him WHERE him.hostid=dhm.hostid ) > 0
                        ELSE true 
                    END)))                                                                              
                  WHERE (dhm.domainid = d.id)) < (ep_min_hosts.val)::integer) OR (5 = ANY (COALESCE(osr.states, '{}'::integer[])))) THEN ARRAY[15]
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN (public.expired_date_test((timezone('MSK'::text, timezone('UTC'::text, d.exdate)))::date, ep_ex_reg.val, ep_tz.val) AND (NOT (2 = ANY (COALESCE(osr.states, '{}'::integer[])))) AND (NOT (1 = ANY (COALESCE(osr.states, '{}'::integer[])))) AND (NOT (35 = ANY (COALESCE(osr.states, '{}'::integer[]))))) THEN ARRAY[17, 2, 4]
            ELSE '{}'::integer[]
        END) AS states
   FROM public.object_registry o,
    ((((((((((public.domain d
     LEFT JOIN public.enumval e ON ((d.id = e.domainid)))
     LEFT JOIN public.object_state_request_now osr ON ((d.id = osr.object_id)))
     JOIN public.enum_parameters ep_ex_not ON ((ep_ex_not.id = 3)))
     JOIN public.enum_parameters ep_ex_dns ON ((ep_ex_dns.id = 4)))
     JOIN public.enum_parameters ep_ex_reg ON ((ep_ex_reg.id = 6)))
     JOIN public.enum_parameters ep_tm ON ((ep_tm.id = 9)))
     JOIN public.enum_parameters ep_tz ON ((ep_tz.id = 10)))
     JOIN public.enum_parameters ep_tm2 ON ((ep_tm2.id = 14)))
     JOIN public.enum_parameters ep_ozu_warn ON ((ep_ozu_warn.id = 18)))
     JOIN public.enum_parameters ep_min_hosts ON ((ep_min_hosts.id = 20)))
  WHERE (d.id = o.id);


CREATE VIEW public.nsset_states AS
 SELECT o.id AS object_id,
    o.historyid AS object_hid,
    ((COALESCE(osr.states, '{}'::integer[]) ||
        CASE
            WHEN ((NOT (d.hostid IS NULL))) THEN ARRAY[16]
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN ((d.hostid IS NULL) AND public.date_month_test(GREATEST((COALESCE(l.last_linked, o.crdate))::date, (COALESCE(ob.update, o.crdate))::date), ep_mn.val, ep_tm.val, ep_tz.val) AND (NOT (1 = ANY (COALESCE(osr.states, '{}'::integer[]))))) THEN ARRAY[17]
            ELSE '{}'::integer[]
        END) AS states
   FROM (((((((public.object ob
     JOIN public.object_registry o ON (((ob.id = o.id) AND (o.type = 2))))
     JOIN public.enum_parameters ep_tm ON ((ep_tm.id = 9)))
     JOIN public.enum_parameters ep_tz ON ((ep_tz.id = 10)))
     JOIN public.enum_parameters ep_mn ON ((ep_mn.id = 11)))
     LEFT JOIN ( SELECT DISTINCT hostid
           FROM public.domain_host_map) d ON ((d.hostid = o.id)))
     LEFT JOIN ( SELECT object_state.object_id,
            max(object_state.valid_to) AS last_linked
           FROM public.object_state
          WHERE (object_state.state_id = 16)
          GROUP BY object_state.object_id) l ON ((o.id = l.object_id)))
     LEFT JOIN public.object_state_request_now osr ON ((o.id = osr.object_id)));

