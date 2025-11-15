const mongoose = require('mongoose');

const ProductSchema = new mongoose.Schema(
  {
    _id: { type: String },
    name: { type: String, required: true, trim: true },
    price: { type: Number, required: true, min: 0 },
    tags: { type: [String], default: [] },
    created_at: { type: Date, required: true },
    updated_at: { type: Date, required: true }
  },
  { versionKey: false }
);

module.exports = mongoose.model('Product', ProductSchema);
