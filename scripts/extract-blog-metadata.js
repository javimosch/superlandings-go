#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

const blogDir = path.join(__dirname, '../sites/vibecode-rescue/v1/blog');

// Get all .md files
const files = fs.readdirSync(blogDir).filter(f => f.endsWith('.md'));

files.forEach(file => {
  const filePath = path.join(blogDir, file);
  const content = fs.readFileSync(filePath, 'utf-8');
  const lines = content.split('\n');

  // Extract title (first h1)
  let title = '';
  let description = '';
  let foundTitle = false;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    
    // Extract h1 title
    if (!foundTitle && line.startsWith('# ')) {
      title = line.substring(2).trim();
      foundTitle = true;
      continue;
    }

    // Extract first paragraph after title as description
    if (foundTitle && description === '' && line.trim() !== '' && !line.startsWith('#')) {
      description = line.trim();
      // Limit description to 150 characters
      if (description.length > 150) {
        description = description.substring(0, 147) + '...';
      }
      break;
    }
  }

  // Create .data.json file
  const dataFile = path.join(blogDir, file + '.data.json');
  const metadata = {
    title: title,
    description: description
  };

  fs.writeFileSync(dataFile, JSON.stringify(metadata, null, 2));
  console.log(`Created ${file}.data.json: "${title}"`);
});

console.log(`\nCreated ${files.length} .data.json files`);
