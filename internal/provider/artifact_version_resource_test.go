package provider

const testAccArtifactVersion = testAcc_ProviderConfig + `
data "vy_artifact_version" "this" {
	application = "petstore-webapp"
}
`

/*
 * TODO: This test doesn't work without a corresponding resource being uploaded beforehand.
 *		 Need to find a way to get this set up without that. I.e. uploading something to S3.
 *		 But in the meanwhile, just remove this comment if you need to test changes and manually upload an artifact.
 */
//func TestAccArtifactVersion_Basic(t *testing.T) {
//	expected_resource_name := "data.vy_artifact_version.this"
//
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		PreCheck:                 func() { testAccPreCheck(t) },
//
//		Steps: []resource.TestStep{
//			{
//				Config: testAccArtifactVersion,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr(expected_resource_name, "application", "petstore-webapp"),
//					resource.TestCheckResourceAttrSet(expected_resource_name, "uri"),
//					resource.TestCheckResourceAttrSet(expected_resource_name, "store"),
//					resource.TestCheckResourceAttrSet(expected_resource_name, "path"),
//					resource.TestCheckResourceAttrSet(expected_resource_name, "version"),
//				),
//			},
//		},
//	})
//}
