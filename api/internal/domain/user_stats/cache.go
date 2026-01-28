package userstats

//deleting the flow
func(s *UserStatsService) PurgeUser(username string){
	s.cache.Delete(username)
}
