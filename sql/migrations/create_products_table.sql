DROP TABLE IF EXISTS products;
CREATE TABLE products (
  id              INT AUTO_INCREMENT NOT NULL,
  title           VARCHAR(100) NOT NULL,
  amount_type     VARCHAR(100) NOT NULL,
  price           DECIMAL(5,2) NOT NULL,
  PRIMARY KEY (`id`)
);