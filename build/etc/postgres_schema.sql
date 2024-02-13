-- Create a sequence for the patients table
CREATE SEQUENCE patients_id_seq START 1;

-- Schema for patients
CREATE TABLE patients (
	id INT PRIMARY KEY DEFAULT nextval('patients_id_seq'),
	name VARCHAR(255) NOT NULL,
	username VARCHAR(255) NOT NULL
);

-- Create a sequence for the images table
CREATE SEQUENCE images_id_seq START 1;

-- Schema for images
CREATE TABLE images (
	id INT PRIMARY KEY DEFAULT nextval('images_id_seq'),
	patient_id INT NOT NULL REFERENCES patients(id),
	bucket_path TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	upload_date TIMESTAMP NOT NULL
);

-- Create a sequence for the tags table
CREATE SEQUENCE tags_id_seq START 1;

-- Schema for tags
CREATE TABLE tags (
	id INT PRIMARY KEY DEFAULT nextval('tags_id_seq'),
	name VARCHAR(255) NOT NULL UNIQUE
);

-- Schema for image_tags
CREATE TABLE image_tags (
	image_id INT NOT NULL REFERENCES images(id),
	tag_id INT NOT NULL REFERENCES tags(id),
	PRIMARY KEY (image_id, tag_id)
);


------------------------------------------------------------------------

-- Populate patients with dummy data
INSERT INTO patients (name, username) VALUES
	('John Doe', 'johndoe'),
	('Jane Smith', 'janesmith'),
	('Michael Johnson', 'michaelj'),
	('Emily Davis', 'emilyd'),
	('Christopher Wilson', 'chrisw');


INSERT INTO images (patient_id, bucket_path, upload_date, name, description) VALUES
																				 (1, '1.jpg', '2024-02-07 08:00:00', 'Image1', 'Description1'),
																				 (2, '2.jpg', '2024-02-08 09:15:25', 'Image2', 'Description2'),
																				 (3, '3.jpg', '2024-02-09 10:30:42', 'Image3', 'Description3'),
																				 (4, '4.jpg', '2024-02-10 11:45:55', 'Image4', 'Description4'),
																				 (5, '5.jpg', '2024-02-11 12:59:59', 'Image5', 'Description5'),
																				 (3, '6.jpeg', '2024-02-12 08:00:00', 'Image6', 'Description6'),
																				 (3, '7.jpeg', '2024-02-13 09:15:25', 'Image7', 'Description7'),
																				 (1, '8.jpeg', '2024-02-14 10:30:42', 'Image8', 'Description8'),
																				 (2, '9.jpeg', '2024-02-15 11:45:55', 'Image9', 'Description9'),
																				 (3, '10.jpeg', '2024-02-16 12:59:59', 'Image10', 'Description10'),
																				 (4, '11.jpeg', '2024-02-17 08:00:00', 'Image11', 'Description11'),
																				 (5, '12.jpeg', '2024-02-18 09:15:25', 'Image12', 'Description12'),
																				 (1, '13.jpeg', '2024-02-19 10:30:42', 'Image13', 'Description13'),
																				 (5, '14.jpeg', '2024-02-20 11:45:55', 'Image14', 'Description14'),
																				 (3, '15.jpeg', '2024-02-21 12:59:59', 'Image15', 'Description15');


-- Populate tags with dummy data
INSERT INTO tags (name) VALUES
	('tag1'),
	('tag2'),
	('tag3'),
	('tag4'),
	('tag5');

-- Populate image_tags with dummy data
INSERT INTO image_tags (image_id, tag_id) VALUES
	(1, 1),
	(1, 2),
	(2, 2),
	(3, 3),
	(4, 4),
	(5, 5),
	(1, 5),
	(1, 4),
	(2, 3),
	(3, 2),
	(4, 1);


-- Index on patients.username
CREATE INDEX idx_patients_username ON patients(username);

-- Index on images.patient_id
CREATE INDEX idx_images_patient_id ON images(patient_id);

-- Index on images.upload_date
CREATE INDEX idx_images_upload_date ON images(upload_date);

-- Indexes on image_tags for both image_id and tag_id
CREATE INDEX idx_image_tags_image_id ON image_tags(image_id);
CREATE INDEX idx_image_tags_tag_id ON image_tags(tag_id);

