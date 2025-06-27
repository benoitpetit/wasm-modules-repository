/**
 * Test local du module jsonxml-wasm
 */

const fs = require('fs');
const path = require('path');

async function testLocal() {
    try {
        console.log('🧪 Test local du module jsonxml-wasm...');
        
        // Charger le fichier WASM local
        const wasmPath = path.join(__dirname, 'jsonxml-wasm', 'main.wasm');
        if (!fs.existsSync(wasmPath)) {
            throw new Error(`Fichier WASM non trouvé: ${wasmPath}`);
        }
        
        console.log(`📁 Chargement du fichier WASM: ${wasmPath}`);
        const wasmBuffer = fs.readFileSync(wasmPath);
        console.log(`📊 Taille du fichier WASM: ${(wasmBuffer.length / 1024 / 1024).toFixed(2)} MB`);
        
        // Charger le script d'exécution Go WASM
        require('./jsonxml-wasm/wasm_exec.js');
        
        // Créer une instance Go
        const go = new Go();
        const wasmModule = await WebAssembly.instantiate(wasmBuffer, go.importObject);
        
        // Exécuter le module
        go.run(wasmModule.instance);
        
        // Attendre que le module soit prêt
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // Tester jsonToXML
        console.log('\n🔄 Test de jsonToXML...');
        const testJson = '{"greeting": "hello", "target": "world"}';
        console.log(`📥 Input JSON: ${testJson}`);
        
        if (typeof global.jsonToXML === 'function') {
            const result = global.jsonToXML(testJson);
            console.log('📤 Résultat:', result);
            console.log('✅ Test réussi !');
        } else {
            console.log('❌ Fonction jsonToXML non disponible');
            console.log('Fonctions disponibles:', Object.keys(global).filter(k => typeof global[k] === 'function' && k.includes('json')));
        }
        
    } catch (error) {
        console.error('❌ Erreur lors du test:', error.message);
        console.error(error.stack);
    }
}

if (require.main === module) {
    testLocal();
}

module.exports = { testLocal }; 