-- Create "projects" table
CREATE TABLE "projects" (
  "id" character varying(255) NOT NULL,
  "name" character varying(255) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create "cells" table
CREATE TABLE "cells" (
  "id" character varying(255) NOT NULL,
  "project_id" character varying(255) NOT NULL,
  "index_position" integer NOT NULL,
  "type" character varying(50) NOT NULL,
  "data" jsonb NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "cells_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_cells_project_id" to table: "cells"
CREATE INDEX "idx_cells_project_id" ON "cells" ("project_id");
-- Create index "idx_cells_project_index" to table: "cells"
CREATE INDEX "idx_cells_project_index" ON "cells" ("project_id", "index_position");
