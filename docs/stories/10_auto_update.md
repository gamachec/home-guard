### Story 10 : Mise à jour automatique (Auto-Update)

**Objectif :** Permettre à l'agent de se mettre à jour automatiquement depuis les releases GitHub, sans aucune intervention de l'utilisateur.

---

#### Contrainte principale

Un processus Windows ne peut pas remplacer son propre exécutable pendant qu'il tourne (le fichier est verrouillé). `home-guard.exe` ne peut donc pas se remplacer lui-même, même s'il arrête le service au préalable : c'est le processus `home-guard.exe update` lui-même qui tient le verrou.

---

#### Solution retenue : second binaire `home-guard-updater.exe`

Un second binaire minimal, `home-guard-updater.exe` (nouveau package `cmd/updater/`), est dédié à la mise à jour. Invoqué par le **Planificateur de tâches Windows**, il peut librement arrêter le service et remplacer `home-guard.exe` puisqu'il s'agit d'un fichier distinct.

---

#### Détails techniques

**1. Versioning des binaires**

La version courante est injectée dans les deux binaires à la compilation via les `ldflags` Go :
```
-ldflags "-X main.version=v1.2.3"
```
Le script Build.ps1, le Makefile et le workflow CI/CD doivent passer le tag de la release en cours pour les deux cibles de build.

---

**2. Fichier `version.txt`**

Au démarrage, `home-guard.exe` écrit sa version dans un fichier `version.txt` situé dans le même répertoire que l'exécutable. Ce fichier est la source de vérité pour `home-guard-updater.exe` afin de connaître la version actuellement installée et éviter toute dérive après une mise à jour.

---

**3. Évolution de `home-guard.exe install` et `uninstall`**

`home-guard.exe install` installe désormais deux choses en une seule commande :
1. Le service Windows `HomeGuard` (comportement existant).
2. Une tâche planifiée Windows nommée `HomeGuardUpdater` qui :
   - Exécute `home-guard-updater.exe` situé dans le même répertoire que `home-guard.exe`
   - S'exécute sous le compte `SYSTEM` pour avoir les droits sur le Service Control Manager
   - Se déclenche au démarrage de la machine puis toutes les heures

`home-guard.exe uninstall` supprime symmétriquement le service et la tâche planifiée.

---

**4. `home-guard-updater.exe` : déroulement**

Lorsqu'il est invoqué par le planificateur :

1. Lit `version.txt` pour obtenir la version installée de `home-guard.exe`.
2. Interroge l'API GitHub REST :
   `https://api.github.com/repos/gamachec/home-guard/releases/latest`
3. Compare le `tag_name` retourné avec la version lue dans `version.txt`.
4. **Si la version distante est identique ou antérieure**, se termine sans rien faire.
5. **Si une mise à jour est disponible** :
   a. Télécharge l'asset `home-guard.exe` vers `home-guard.exe.new` dans le répertoire de l'exécutable.
   b. Télécharge `checksums.txt` et vérifie le hash SHA-256 de `home-guard.exe.new`. En cas de mismatch, supprime `home-guard.exe.new` et abandonne avec une erreur dans les logs.
   c. Arrête le service `HomeGuard` via le Service Control Manager et attend sa terminaison complète.
   d. Remplace `home-guard.exe` par `home-guard.exe.new`.
   e. Redémarre le service `HomeGuard`.
   f. Supprime `home-guard.exe.new` s'il est encore présent.
   g. En cas d'erreur après l'arrêt du service, tente de relancer le service et logue l'erreur (le binaire original est conservé si le remplacement n'a pas eu lieu).

---

**5. Publication de la version sur MQTT**

Au démarrage, `home-guard.exe` publie sa version sur :
- **Topic :** `stat/<client_id>/version`
- **Payload :** `v1.2.3` (valeur string brute)

---

**6. MQTT Discovery : capteur de version**

Ajouter dans le payload de discovery au démarrage :

- **Topic :** `homeassistant/sensor/<client_id>/version/config`
- **Payload :**
```json
{
  "name": "Version",
  "unique_id": "<client_id>_version",
  "state_topic": "stat/<client_id>/version",
  "device": { "identifiers": ["<client_id>"] }
}
```

---

**7. Release GitHub : assets requis**

Chaque release GitHub doit publier trois assets :
- `home-guard.exe` — le binaire principal compilé pour Windows amd64
- `home-guard-updater.exe` — le binaire updater compilé pour Windows amd64
- `checksums.txt` — hashes SHA-256 des deux exécutables, au format standard `sha256sum`

Le workflow CI/CD (`release.yml`) doit être modifié pour builder les deux binaires et uploader les trois fichiers.

> **Note :** `home-guard-updater.exe` n'est pas mis à jour automatiquement (il ne peut pas se remplacer lui-même). Étant minimal et stable, il est mis à jour manuellement lors d'une réinstallation (`home-guard.exe install`). La commande `install` peut être relancée sans désinstaller au préalable.
