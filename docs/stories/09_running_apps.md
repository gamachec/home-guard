# Story 9 : amélioration running apps

Je veux modifier le comportement de ce qui est exposé sur le topic stat/<client>/running_apps récupéré aujourd'hui
depuis la méthode RunningFromBlacklist dans internal/process/manager.go:69.

Je veux désormais avoir la liste des applications qui tournent en tant qu'Applications pour l'utilisateur courant,
Windows semble faire la distinction entre les "Applications" et les "Processus en arrière plan" quand on affiche la
liste des Processus, il faut chercher comment il fait cette distinction.

Je veux également, pour chaque process, avoir son nom, son emplacement sur le disque, ainsi que la description du
fichier (disponible dans ses métadonnées) dans la liste "running_apps" exposée.
