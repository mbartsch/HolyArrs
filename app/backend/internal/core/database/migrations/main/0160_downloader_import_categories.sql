ALTER TABLE downloaders ADD COLUMN category_after_import_movies TEXT NOT NULL DEFAULT '';
ALTER TABLE downloaders ADD COLUMN category_after_import_shows TEXT NOT NULL DEFAULT '';
ALTER TABLE downloaders ADD COLUMN category_after_import_music TEXT NOT NULL DEFAULT '';
