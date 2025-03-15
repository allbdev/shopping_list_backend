-- Create lists table
CREATE TABLE lists (
    id INT AUTO_INCREMENT PRIMARY KEY,
    workspace_id INT NOT NULL,
    user_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create list_products table to store the products in each list
CREATE TABLE list_products (
    list_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT DEFAULT 1,
    checked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (list_id) REFERENCES lists(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    PRIMARY KEY (list_id, product_id)
);

-- Add indexes for better query performance
CREATE INDEX idx_lists_workspace ON lists(workspace_id);
CREATE INDEX idx_lists_user ON lists(user_id);
CREATE INDEX idx_list_products_list ON list_products(list_id);
CREATE INDEX idx_list_products_product ON list_products(product_id); 