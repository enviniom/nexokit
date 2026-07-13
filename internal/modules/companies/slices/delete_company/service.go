package delete_company

type Service interface{ Delete(publicID string) error }
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Delete(id string) error {
	if _, err := s.repo.GetByPublicID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
