import { readFileSync, readdirSync } from 'node:fs';
import { dirname, extname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const frontendRoot = dirname(dirname(fileURLToPath(import.meta.url)));
const appRoot = join(frontendRoot, 'src', 'app');
const iconRoot = join(frontendRoot, 'src', 'assets', 'icons');
const namespaces = ['outline', 'solid'];

function sourceFiles(directory) {
    return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
        const path = join(directory, entry.name);

        if (entry.isDirectory()) {
            return sourceFiles(path);
        }

        return ['.html', '.ts'].includes(extname(entry.name)) ? [path] : [];
    });
}

const availableIcons = new Map(
    namespaces.map((namespace) => {
        const sprite = readFileSync(join(iconRoot, `heroicons-${namespace}.svg`), 'utf8');
        return [namespace, new Set(Array.from(sprite.matchAll(/\bid="([^"]+)"/g), (match) => match[1]))];
    })
);

const missingReferences = [];

for (const path of sourceFiles(appRoot)) {
    const source = readFileSync(path, 'utf8');

    for (const match of source.matchAll(/heroicons_(outline|solid):([a-z0-9-]+)/g)) {
        const [, namespace, icon] = match;

        if (!availableIcons.get(namespace).has(icon)) {
            const line = source.slice(0, match.index).split('\n').length;
            missingReferences.push(`${path.slice(frontendRoot.length + 1)}:${line} heroicons_${namespace}:${icon}`);
        }
    }
}

if (missingReferences.length > 0) {
    console.error('Heroicons references missing from registered SVG bundles:');
    console.error(missingReferences.join('\n'));
    process.exitCode = 1;
} else {
    console.log('All Heroicons references resolve to registered SVG symbols.');
}
