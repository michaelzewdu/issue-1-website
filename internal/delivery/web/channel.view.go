package web

import (
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"net/http"
)

func getChannelView(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := getParametersFromRequestAsMap(r)
		channelUsername := vars["channelUsername"]

		sess, err := SessionStartLoggedIn(s, w, r)
		if err != nil {
			return
		}
		var channelData struct {
			posts            []*issue1.Post
			Releases         []*issue1.Release
			OfficialReleases []*issue1.Release
			Admins           []string
			Owner            string
			*NavBarData
		}
		channelData.NavBarData, err = getNavbarData(s, sess, w, r)
		if err != nil {
			return
		}
		authToken := sess.Get(s.sessionValues.restRefreshToken)
		channelData.posts, err = s.Iss1C.ChannelService.GetChannelPosts(channelUsername)
		s.Logger.Printf("here2")
		if err != nil {
			if err == issue1.ErrPostNotFound {
				s.Logger.Printf("here1")
				show404Page(w, r)
				return
			}
			s.Logger.Printf("here3")
			showErrorPage(w, r)
			return
		}
		channelData.Releases = make([]*issue1.Release, 0)
		//rel,err:=s.Iss1C.ChannelService.GetCatalog(channelUsername,authToken)
		//s.Logger.Printf("here4")
		//if err != nil {
		//	s.Logger.Printf(err.Error())
		//	return
		//}
		//s.Logger.Printf("here46")
		//for _, release := range rel {
		//	channelData.Releases = append(channelData.Releases, release)
		//}
		//s.Logger.Printf("here90")
		channelData.OfficialReleases = make([]*issue1.Release, 0)
		//rel,err :=s.Iss1C.ChannelService.GetOfficialCatalog(channelUsername,authToken)
		//if err != nil {
		//	s.Logger.Printf(err.Error())
		//	return
		//}
		//for _, release := range rel {
		//	channelData.OfficialReleases = append(channelData.OfficialReleases, release)
		//}
		channelData.Admins = make([]string, 0)
		adm, err := s.Iss1C.ChannelService.GetAdmins(channelUsername, authToken)
		if err != nil {
			return
		}
		for _, user := range adm {
			channelData.Admins = append(channelData.Admins, user)
		}
		cha, err := s.Iss1C.ChannelService.GetOwner(channelUsername, authToken)
		if err != nil {
			return
		}
		channelData.Owner = cha
		s.Logger.Printf("herefinal")
		_ = s.templates.ExecuteTemplate(w, "channel.view", channelData)
	}
}
