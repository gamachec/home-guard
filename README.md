# Home Guard

Agent de supervision et contrôle parental pour Windows, piloté par Home Assistant via MQTT.

## Prérequis

- Un broker MQTT accessible (Mosquitto, EMQX, etc.)
- Home Assistant connecté au même broker

## Configuration

Copiez `config.example.json` en `config.json` dans le même répertoire que l'exécutable :

```json
{
  "broker": "192.168.1.10",
  "port": 1883,
  "username": "user",
  "password": "secret",
  "client_id": "pc-enfant",
  "blacklist": ["roblox.exe", "fortnite.exe", "discord.exe"]
}
```

| Champ       | Description                                                              |
|-------------|--------------------------------------------------------------------------|
| `broker`    | Adresse IP ou hostname du broker MQTT                                    |
| `port`      | Port du broker (défaut : `1883`)                                         |
| `username`  | Identifiant MQTT (laisser vide si sans authentification)                 |
| `password`  | Mot de passe MQTT (laisser vide si sans authentification)                |
| `client_id` | Identifiant unique de cet agent, utilisé dans tous les topics MQTT       |
| `blacklist` | Liste des exécutables à surveiller et fermer de force en mode `BLOCKED`  |

> La blacklist peut être mise à jour dynamiquement depuis Home Assistant sans redémarrer l'agent.

## Installation du service Windows

L'agent peut s'exécuter en tant que service Windows (démarrage automatique, tâche de fond invisible) ou
directement en ligne de commande pour le débogage.

### Installer le service

```powershell
home-guard.exe install
```

Le service `HomeGuard` est enregistré avec le démarrage automatique.
Les logs sont écrits dans `dist/service.log`.

### Désinstaller le service

```powershell
home-guard.exe uninstall
```

## Topics MQTT

L'agent publie et écoute les topics suivants (remplacer `<client_id>` par la valeur de la config) :

| Topic                              | Direction | Description                                      |
|------------------------------------|-----------|--------------------------------------------------|
| `stat/<client_id>/status`          | Publication | `online` ou `offline` (LWT automatique)        |
| `stat/<client_id>/current_mode`    | Publication | Mode actif : `ACTIVE` ou `BLOCKED`             |
| `stat/<client_id>/running_apps`    | Publication | Tableau JSON des apps blacklistées en cours     |
| `cmnd/<client_id>/mode`            | Réception | Changer le mode : `ACTIVE` ou `BLOCKED`         |
| `cmnd/<client_id>/notify`          | Réception | Afficher une notification Windows (JSON)         |
| `cmnd/<client_id>/blacklist/set`   | Réception | Mettre à jour la blacklist (tableau JSON)        |

## Entités Home Assistant (auto-discovery)

L'agent publie automatiquement sa configuration à chaque connexion via le mécanisme d'auto-discovery de
Home Assistant. Aucune configuration manuelle n'est nécessaire dans Home Assistant.

### Sélecteur de mode

**Type :** `select`

Permet de basculer entre les modes `ACTIVE` et `BLOCKED` directement depuis le tableau de bord Home
Assistant ou dans des automatisations.

- En mode `ACTIVE` : surveillance passive uniquement.
- En mode `BLOCKED` : les applications de la blacklist sont détectées et fermées de force toutes les secondes.

### Capteur de connectivité

**Type :** `binary_sensor` — classe `connectivity`

Indique si l'agent est en ligne (`ON`) ou hors ligne (`OFF`). Le passage à `OFF` est automatique grâce
au mécanisme MQTT Last Will and Testament : Home Assistant est notifié même en cas de coupure réseau
ou de crash.

Utilisable dans les automatisations pour détecter que le PC est allumé/éteint.

### Capteur des applications en cours

**Type :** `sensor`

Affiche en temps réel les applications de la blacklist actuellement en cours d'exécution sur le PC,
sous forme de tableau JSON (ex : `["roblox.exe", "discord.exe"]`).

Utile pour surveiller l'activité et déclencher des automatisations (ex : envoyer une notification
aux parents si une application interdite est lancée en mode `ACTIVE`).