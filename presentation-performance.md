# Comparaison Performance: Node.js HTTP vs Go gRPC
## Transfert de Données Inter-Services

---

## Diapositive 1: Description de la Comparaison

### 🎯 **Objectif**
Évaluer les performances de transfert de données entre microservices sous charge concurrente

### 🏗️ **Architecture Testée**

**Stack Node.js HTTP:**
- Gateway Express.js (port 3000) → Microservice Express.js (port 3001)
- Communication via REST API (`fetch()`)
- Transfert JSON direct

**Stack Go gRPC:**
- Gateway Go/Gin (port 8080) → Microservice gRPC (port 50051)
- Communication via Protocol Buffers
- Streaming gRPC avec chunks configurables

### 📊 **Conditions de Test**
- **Dataset:** 1000 hôtels (900 disponibles)
- **Concurrence:** 10 appels simultanés
- **Métrique:** Temps de traitement par appel et temps total

---

## Diapositive 2: Résultats de Performance

### 📈 **Résultats Comparatifs**

| Métrique | Node.js HTTP | Go gRPC | **Amélioration** |
|----------|--------------|---------|------------------|
| **Temps Total** | 4,642 ms | 1,565 ms | **🚀 2.97x plus rapide** |
| **Temps Moyen** | 3,650 ms | 1,526 ms | **⚡ 2.39x plus rapide** |
| **Temps Min** | 2,660 ms | 1,475 ms | **🏃 1.80x plus rapide** |
| **Temps Max** | 4,641 ms | 1,565 ms | **🎯 2.97x plus rapide** |
| **Variance** | Haute (±1,981 ms) | Faible (±90 ms) | **📊 95% plus stable** |

### 🏆 **Conclusions Clés**

✅ **gRPC** offre des performances **~3x supérieures** pour le transfert de données  
✅ **Stabilité** remarquable: variance réduite de 95%  
✅ **Scalabilité** optimale sous charge concurrente  
✅ **Protocol Buffers** + streaming = efficacité maximale  

### ⚠️ **Faiblesses du Stack Node.js HTTP**

❌ **Temps total excessif** : 4,642ms vs 1,565ms (gRPC) - **66% plus lent**  
❌ **Overhead cumulé** : chaque appel concurrent pénalise le temps total  
❌ **Saturation réseau** : JSON volumineux ralentit l'ensemble des transferts  
❌ **Goulot d'étranglement** : traitement séquentiel vs streaming parallèle  
❌ **Inefficacité globale** : accumulation des latences individuelles  