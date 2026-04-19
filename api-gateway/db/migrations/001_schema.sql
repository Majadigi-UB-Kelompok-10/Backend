-- Service List Table --
CREATE TABLE IF NOT EXISTS "public"."service_list" (
    "service_list_id"   "uuid"                      DEFAULT "gen_random_uuid"() NOT NULL,
    "title"             "text"                      UNIQUE                      NOT NULL,
    "description"       "text"                                                          ,
    "icon_url"          "text"                                                  NOT NULL,
    "created_at"        timestamp with time zone    DEFAULT "now"()             NOT NULL
);

COMMENT ON TABLE "public"."service_list" IS 'Services available for Majadigi';

ALTER TABLE ONLY "public"."service_list"
    ADD CONSTRAINT "service_list_pkey" PRIMARY KEY ("service_list_id");

-- Category List Table --
CREATE TABLE IF NOT EXISTS "public"."category_list" (
    "category_list_id"  "uuid"                      DEFAULT "gen_random_uuid"() NOT NULL,
    "name"              "text"                      UNIQUE                      NOT NULL,
    "description"       "text"                                                          ,
    "created_at"        timestamp with time zone    DEFAULT "now"()             NOT NULL
);

COMMENT ON TABLE "public"."category_list" IS 'List of Categories available for Services';

ALTER TABLE ONLY "public"."category_list"
    ADD CONSTRAINT "category_list_pkey" PRIMARY KEY ("category_list_id");

-- Services has Categories Table (Many-to-Many) --
CREATE TABLE IF NOT EXISTS "public"."services_has_categories" (
    "service_list_id"   "uuid"                                                  NOT NULL,
    "category_list_id"  "uuid"                                                  NOT NULL,
    "created_at"        timestamp with time zone    DEFAULT "now"()             NOT NULL
);

ALTER TABLE ONLY "public"."services_has_categories"
    ADD CONSTRAINT "service_list_id_fkey" FOREIGN KEY ("service_list_id") REFERENCES "public"."service_list"("service_list_id") ON DELETE CASCADE,
    ADD CONSTRAINT "category_list_id_fkey" FOREIGN KEY ("category_list_id") REFERENCES "public"."category_list"("category_list_id") ON DELETE CASCADE;

-- Endpoint List Table --
CREATE TABLE IF NOT EXISTS "public"."endpoint_list" (
    "endpoint_list_id"  "uuid"                      DEFAULT "gen_random_uuid"() NOT NULL,
    "slug_name"         "text"                      UNIQUE                      NOT NULL,
    "page_url"          "text"                                                  NOT NULL,
    "created_at"        timestamp with time zone    DEFAULT "now"()             NOT NULL
);

COMMENT ON TABLE "public"."endpoint_list" IS 'Registry of Endpoint available in Backend';

ALTER TABLE ONLY "public"."endpoint_list"
    ADD CONSTRAINT "endpoint_list_pkey" PRIMARY KEY ("endpoint_list_id");

-- Image List Table -- 
CREATE TABLE IF NOT EXISTS "public"."image_list" (
    "image_list_id"     "uuid"                      DEFAULT "gen_random_uuid"() NOT NULL,
    "service_list_id"   "uuid"                                                  NOT NULL,
    "image_url"         "text"                                                  NOT NULL,
    "semantic_label"    "text"                                                          , 
    "created_at"        timestamp with time zone    DEFAULT "now"()             NOT NULL
);

COMMENT ON TABLE "public"."image_list" IS 'Images assets for each services';

ALTER TABLE ONLY "public"."image_list"
    ADD CONSTRAINT "image_list_pkey" PRIMARY KEY ("image_list_id"),
    ADD CONSTRAINT "service_list_id_fkey" FOREIGN KEY ("service_list_id") REFERENCES "public"."service_list"("service_list_id") ON DELETE CASCADE;

-- Integration List Table --
CREATE TABLE IF NOT EXISTS "public"."integration_list" (
    "integration_list_id"   "uuid"                      DEFAULT "gen_random_uuid"() NOT NULL,
    "service_list_id"       "uuid"                                                  NOT NULL,
    "endpoint_list_id"      "uuid"                                                  NOT NULL,
    "title"                 "text"                                                  NOT NULL,
    "icon_url"              "text"                                                          ,
    "created_at"            timestamp with time zone    DEFAULT "now"()             NOT NULL
);

COMMENT ON TABLE "public"."integration_list" IS 'Per-service integration list';

ALTER TABLE ONLY "public"."integration_list"
    ADD CONSTRAINT "integration_list_pkey" PRIMARY KEY ("integration_list_id"),
    ADD CONSTRAINT "service_list_id_fkey" FOREIGN KEY ("service_list_id") REFERENCES "public"."service_list"("service_list_id") ON DELETE CASCADE,
    ADD CONSTRAINT "endpoint_list_id_fkey" FOREIGN KEY ("endpoint_list_id") REFERENCES "public"."endpoint_list"("endpoint_list_id") ON DELETE CASCADE;

-- Operational List Table --
CREATE TABLE IF NOT EXISTS "public"."operational_list" (
    "operational_list_id"   "uuid"                      DEFAULT "gen_random_uuid"() NOT NULL,
    "service_list_id"       "uuid"                      UNIQUE                      NOT NULL,
    "service_url"           "text"                                                  NOT NULL,
    "address"               "text"                                                          ,
    "operational_hour"      "jsonb"                                                         ,
    "social_media"          "jsonb"                                                         ,
    "created_at"            timestamp with time zone    DEFAULT "now"()             NOT NULL
);

COMMENT ON TABLE "public"."operational_list" IS 'Data for operational tab';

ALTER TABLE ONLY "public"."operational_list"
    ADD CONSTRAINT "operational_list_pkey" PRIMARY KEY ("operational_list_id"),
    ADD CONSTRAINT "service_list_id_fkey" FOREIGN KEY ("service_list_id") REFERENCES "public"."service_list"("service_list_id") ON DELETE CASCADE;

-- Policy List Table --
CREATE TABLE IF NOT EXISTS "public"."policy_list" (
    "policy_list_id"    "uuid"                      DEFAULT "gen_random_uuid"() NOT NULL,
    "service_list_id"   "uuid"                      UNIQUE                      NOT NULL,
    "benefit"           "jsonb"                                                         ,
    "instruction"       "jsonb"                                                         ,
    "created_at"        timestamp with time zone    DEFAULT "now"()             NOT NULL
);

COMMENT ON TABLE "public"."policy_list" IS 'Policies data for services';

ALTER TABLE ONLY "public"."policy_list"
    ADD CONSTRAINT "policy_list_pkey" PRIMARY KEY ("policy_list_id"),
    ADD CONSTRAINT "service_list_id_fkey" FOREIGN KEY ("service_list_id") REFERENCES "public"."service_list"("service_list_id") ON DELETE CASCADE;



CREATE OR REPLACE FUNCTION notify_gateway_update()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM pg_notify('gateway_refresh_channel', 'refresh_now');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_endpoint_update ON endpoint_list;

CREATE TRIGGER trg_endpoint_update
AFTER INSERT OR UPDATE OR DELETE ON endpoint_list
FOR EACH STATEMENT EXECUTE FUNCTION notify_gateway_update();