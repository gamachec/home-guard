# Project Overview

**Objectif Global :**
Développer un agent d'arrière-plan léger pour Windows, écrit en **Go**, permettant de superviser et de restreindre le
temps d'utilisation d'un ordinateur enfant via une intégration domotique (Home Assistant).

**Fonctionnalités Clés :**

1. **Contrôle d'Exécution :** Surveillance active des processus Windows et fermeture forcée des applications
   interdites (jeux, navigateurs) selon l'état de l'agent.
2. **Système de Modes :** Gestion de trois états de fonctionnement :
    * `ACTIVE` : Utilisation libre.
    * `WARNING` : Envoi d'une notification native Windows (Toast) pour prévenir d'une coupure imminente (ex: J-5 min).
    * `BLOCKED` : Application stricte de la "Blacklist" avec scan fréquent et terminaison immédiate des processus
      interdits.
3. **Communication Pilotée par Événements :** Utilisation du protocole **MQTT** pour une interaction bidirectionnelle en
   temps réel avec Home Assistant (réception de commandes et publication d'états).
4. **Discrétion et Performance :** Exécution en tant que service ou processus d'arrière-plan invisible, avec une
   empreinte mémoire minimale (< 20 Mo RAM).

**Stack Technique :**

* **Langage :** Go (Golang) pour sa compilation statique en un `.exe` unique et sa gestion efficace de la concurrence (
  goroutines).
* **Communication :** Client MQTT (Paho) avec support du "Last Will and Testament" pour le monitoring de présence.
* **OS Interop :** Appels système Windows (Win32 API) pour le listing/kill de processus et l'affichage des notifications
  système.
* **Configuration :** Fichier local persistant (JSON/YAML) pour la blacklist et les paramètres de connexion.

**Philosophie de Conception :**
L'agent doit être résilient (reconnexion automatique au broker MQTT) et sécurisé (ne traiter que les commandes provenant
du broker autorisé). Il ne contient pas de logique de planning complexe ; celle-ci est déportée dans Home Assistant,
l'agent agissant comme un exécuteur d'ordres et un capteur d'état.

# Règles de développement OBLIGATOIRES

- Ne pas commenter le code, préferer un code lisible.
- Ne pas générer de documentation, à moins que ce soit explicitement demandé
- Applique les principes de programmation orientée objet à la mode Go : isole chaque fonctionnalité dans des structs
  dédiées avec leurs propres méthodes. Utilise des constructeurs pour l'injection de dépendances et des interfaces pour
  découpler les modules (MQTT, ProcessManager, Notifier...). Le code doit être modulaire, sans variables globales, et
  chaque service doit être géré via des goroutines et des Contexts pour un arrêt propre.
- Rédige systématiquement des tests unitaires pour chaque nouvelle fonctionnalité (fichiers _test.go). Concentre-toi
  uniquement sur les 'happy paths' (cas passants) pour valider le comportement nominal, sans chercher une couverture
  exhaustive des cas d'erreur.

Inutile d'essayer de lancer les commandes GO dans le terminal, tu tourne dans un WSL et GO est installé sur mon
filesystem Windows.